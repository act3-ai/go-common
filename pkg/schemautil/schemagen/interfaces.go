package schemagen

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// SchemaExtender is implemented by types that want to modify their
// generated JSON Schema representation.
type SchemaExtender interface {
	// ExtendJSONSchema modifies the generated JSON Schema representation of a type.
	ExtendJSONSchema(schema *jsonschema.Schema)
}

// SchemaProvider is implemented by types that want to directly provide
// their JSON Schema representation.
type SchemaProvider interface {
	// JSONSchema produces the JSON Schema representation of a type.
	JSONSchema() *jsonschema.Schema
}

var (
	typeSchemaExtender = reflect.TypeFor[SchemaExtender]()
	typeSchemaProvider = reflect.TypeFor[SchemaProvider]()
)
