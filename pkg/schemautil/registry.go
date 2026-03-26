package schemautil

import (
	"github.com/google/jsonschema-go/jsonschema"
)

// Registry defines a registry of JSON Schemas.
type Registry interface {
	// GetSchema returns the JSON Schema at the given reference.
	GetSchema(ref string) (*jsonschema.Schema, bool)
}

var _ Registry = MapRegistry(nil)

// MapRegistry is a Registry in a map.
type MapRegistry map[string]*jsonschema.Schema

// GetSchema implements Registry.
func (reg MapRegistry) GetSchema(ref string) (*jsonschema.Schema, bool) {
	if len(reg) == 0 {
		return nil, false
	}
	schema, ok := reg[ref]
	return schema, ok
}
