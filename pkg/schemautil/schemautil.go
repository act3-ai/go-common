// package schemautil contains utilities for working with JSON Schemas.
package schemautil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
	"slices"

	"github.com/google/jsonschema-go/jsonschema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// JSON Schema type values.
const (
	TypeArray   = "array"
	TypeBoolean = "boolean"
	TypeInteger = "integer"
	TypeNull    = "null"
	TypeNumber  = "number"
	TypeObject  = "object"
	TypeString  = "string"
)

// JSON Schema format values.
const (
	FormatDate     = "date"
	FormatDateTime = "date-time"
	FormatTime     = "time"
	FormatInt32    = "int32"
	FormatInt64    = "int64"
	FormatFloat    = "float"
	FormatDouble   = "double"
)

// OrderedProperties returns an iterator over the properties of a
// schema in their defined order, if they are defined with an order.
func OrderedProperties(schema *jsonschema.Schema) iter.Seq2[string, *jsonschema.Schema] {
	return func(yield func(string, *jsonschema.Schema) bool) {
		for propName := range OrderedPropertyNames(schema) {
			if !yield(propName, schema.Properties[propName]) {
				return
			}
		}
	}
}

// OrderedPropertyNames returns an iterator over the property names of a
// schema in their defined order, if they are defined with an order.
func OrderedPropertyNames(schema *jsonschema.Schema) iter.Seq[string] {
	return func(yield func(string) bool) {
		if schema == nil || schema.Properties == nil {
			return
		}

		for _, propName := range schema.PropertyOrder {
			if !yield(propName) {
				return
			}
		}

		// Lookup for ordered property names
		orderedProps := sets.New(schema.PropertyOrder...)

		// Create list of remaining property names
		remainingProperties := make([]string, 0, len(schema.Properties)-len(schema.PropertyOrder))
		for propName := range schema.Properties {
			if !orderedProps.Has(propName) {
				// Add property if it is not in PropertyOrder
				remainingProperties = append(remainingProperties, propName)
			}
		}

		// Iterate over the remaining properties in deterministic order
		for _, propName := range slices.Sorted(slices.Values(remainingProperties)) {
			if !yield(propName) {
				return
			}
		}
	}
}

// TrueSchema returns a Schema that validates any JSON value.
// It is equivalent to the empty schema ({}), which marshals to the JSON literal true.
func TrueSchema() *jsonschema.Schema {
	return &jsonschema.Schema{}
}

// FalseSchema returns a Schema that validates no JSON value.
// It is equivalent to the schema {"not": {}}, which marshals to the JSON literal false.
func FalseSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Not: TrueSchema()}
}

// IsTrueSchema reports whether the schema is the true schema, as defined by TrueSchema.
func IsTrueSchema(schema *jsonschema.Schema) bool {
	data, _ := json.Marshal(schema) //nolint:errchkjson // if this fails, the schema is not the true schema
	return bytes.Equal(data, []byte(`true`))
}

// IsFalseSchema reports whether the schema is the false schema, as defined by FalseSchema.
func IsFalseSchema(schema *jsonschema.Schema) bool {
	data, _ := json.Marshal(schema) //nolint:errchkjson // if this fails, the schema is not the false schema
	return bytes.Equal(data, []byte(`false`))
}

// SetExtension sets an extension in the schema.
func SetExtension(schema *jsonschema.Schema, key string, value any) {
	if schema == nil {
		panic("nil schema")
	}
	// Initialize extras map if needed
	if schema.Extra == nil {
		schema.Extra = make(map[string]any, 1)
	}
	// Set the extension
	schema.Extra[key] = value
	// Nest reference in an allOf schema
	NestReference(schema)
}

// NestReference modifies the schema to nest the $ref keyword in its own subschema if necessary.
func NestReference(schema *jsonschema.Schema) {
	if schema == nil || schema.Ref == "" {
		return
	}

	// Store the reference
	ref := schema.Ref

	// Overwrite with empty string
	schema.Ref = ""

	// If schema is empty without the reference,
	// nesting is not necessary
	if IsTrueSchema(schema) {
		schema.Ref = ref
		return
	}

	// Move the reference to a subschema in the allOf list
	schema.AllOf = append(schema.AllOf, &jsonschema.Schema{
		Ref: ref,
	})
}

// ReachableRefs collects all $ref values that can be reached by walking the schema
// and following any contained references.
func ReachableRefs(reg Registry, schema *jsonschema.Schema) ([]string, error) {
	visited := sets.New[string]()
	if err := visitRefs(visited, reg, schema); err != nil {
		return nil, err
	}
	return visited.UnsortedList(), nil
}

func visitRefs(visited sets.Set[string], reg Registry, schema *jsonschema.Schema) error {
	_, err := WalkSchema(schema, func(loc string, schema *jsonschema.Schema) (*jsonschema.Schema, error) {
		if schema == nil || schema.Ref == "" {
			return schema, nil
		}

		// Referenced schema already visited
		if visited.Has(schema.Ref) {
			return schema, nil
		}

		// Add to visited
		visited.Insert(schema.Ref)

		// Get the schema
		next, ok := reg.GetSchema(schema.Ref)
		if !ok {
			return schema, fmt.Errorf("%s/$ref: no schema matching %q", loc, schema.Ref)
		}

		// Walk the referenced schema
		return schema, visitRefs(visited, reg, next)
	})
	return err
}
