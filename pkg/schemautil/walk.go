package schemautil

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/act3-ai/go-common/pkg/jsonpointer"
)

// WalkSchemaFunc is called for each subschema node while walking a JSON Schema.
type WalkSchemaFunc func(loc string, schema *jsonschema.Schema) (*jsonschema.Schema, error)

// WalkSchema walks all subschemas of a JSON Schema.
func WalkSchema(in *jsonschema.Schema, cb WalkSchemaFunc) (*jsonschema.Schema, error) {
	return WalkSchemaWithLocation("", in, cb)
}

// WalkSchemaWithLocation walks all subschemas of a JSON Schema.
func WalkSchemaWithLocation(loc string, in *jsonschema.Schema, cb WalkSchemaFunc) (*jsonschema.Schema, error) {
	// Do not recurse if schema is nil
	if in == nil {
		return in, nil
	}

	slog.Debug("walking schema", slog.String("loc", loc))

	// Run callback
	in, err := cb(loc, in)
	if err != nil {
		return in, fmt.Errorf("%s: %w", loc, err)
	}

	// Run callbacks for all subschemas

	err = walkSubschemaMap(loc+"/properties", in.Properties, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaMap(loc+"/$defs", in.Defs, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaMap(loc+"/definitions", in.Definitions, cb)
	if err != nil {
		return in, err
	}
	in.AdditionalProperties, err = WalkSchemaWithLocation(loc+"/additionalProperties", in.AdditionalProperties, cb)
	if err != nil {
		return in, err
	}
	in.PropertyNames, err = WalkSchemaWithLocation(loc+"/propertyNames", in.PropertyNames, cb)
	if err != nil {
		return in, err
	}
	in.Not, err = WalkSchemaWithLocation(loc+"/not", in.Not, cb)
	if err != nil {
		return in, err
	}
	in.Items, err = WalkSchemaWithLocation(loc+"/items", in.Items, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaList(loc+"/items", in.ItemsArray, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaMap(loc+"/patternProperties", in.PatternProperties, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaList(loc+"/allOf", in.AllOf, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaList(loc+"/anyOf", in.AnyOf, cb)
	if err != nil {
		return in, err
	}
	err = walkSubschemaList(loc+"/oneOf", in.OneOf, cb)
	if err != nil {
		return in, err
	}
	return in, nil
}

// Walks a list of subschemas.
func walkSubschemaList(loc string, subschemas []*jsonschema.Schema, cb WalkSchemaFunc) error {
	for i, subschema := range subschemas {
		modified, err := WalkSchemaWithLocation(loc+"/"+strconv.Itoa(i), subschema, cb)
		if err != nil {
			return err
		}
		subschemas[i] = modified
	}
	return nil
}

// Walks a map of subschemas.
func walkSubschemaMap(loc string, subschemas map[string]*jsonschema.Schema, cb WalkSchemaFunc) error {
	for subkey, subschema := range subschemas {
		modified, err := WalkSchemaWithLocation(loc+"/"+jsonpointer.Escape(subkey), subschema, cb)
		if err != nil {
			return err
		}
		subschemas[subkey] = modified
	}
	return nil
}
