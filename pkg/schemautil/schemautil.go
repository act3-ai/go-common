// package schemautil contains utilities for working with JSON Schemas.
package schemautil

import (
	"bytes"
	"encoding/json"
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

// JSON Schema content encoding values.
// https://json-schema.org/draft/2020-12/draft-bhutton-json-schema-validation-01#section-8.3
const (
	ContentEncodingQuotedPrintable = "quoted-printable"
	ContentEncodingBase16          = "base16"
	ContentEncodingBase32          = "base32"
	ContentEncodingBase64          = "base64"
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
