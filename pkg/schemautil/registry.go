package schemautil

import (
	"errors"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	// ErrSchemaNotFound is returned when a schema is not found.
	ErrSchemaNotFound = errors.New("schema not found")
)

// Registry defines schema retrieval.
type Registry interface {
	// GetSchema returns the schema at the given reference.
	GetSchema(ref string) *jsonschema.Schema
}

var _ Registry = MapRegistry(nil)

// MapRegistry is a Registry in a map.
type MapRegistry map[string]*jsonschema.Schema

// GetSchema implements Registry.
func (reg MapRegistry) GetSchema(ref string) *jsonschema.Schema {
	if len(reg) == 0 {
		return nil
	}
	return reg[ref]
}
