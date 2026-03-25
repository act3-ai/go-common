package schemautil

import (
	"cmp"
	"log/slog"
	"slices"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/iancoleman/orderedmap"

	"github.com/act3-ai/go-common/pkg/jsonpointer"
)

// NewExampleGenerator creates a new ExampleGenerator for a schema registry.
func NewExampleGenerator(reg Registry) *ExampleGenerator {
	return &ExampleGenerator{
		reg:     reg,
		results: map[string]result{},
	}
}

// ExampleGenerator generates example data from JSON Schemas.
type ExampleGenerator struct {
	reg     Registry          // schema registry
	results map[string]result // cached results
}

// result from generating example data.
type result struct {
	example any  // the example data
	ok      bool // success indicator
}

// GetExample gets the example value for the reference.
func (gen *ExampleGenerator) GetExample(ref string) (any, bool) {
	if _, ok := gen.results[ref]; !ok {
		// Generate example if not generated yet
		example, ok := gen.generateSchemaExample(ref, gen.reg.GetSchema(ref))
		gen.results[ref] = result{example, ok}
	}
	// Return generated example
	r := gen.results[ref]
	return r.example, r.ok
}

func (gen *ExampleGenerator) generateSchemaExample(loc string, schema *jsonschema.Schema) (any, bool) { //nolint:gocognit
	// Empty value, dead end
	if schema == nil {
		slog.Error("empty schema", slog.String("location", loc))
		return nil, false
	}
	// Schema references another schema
	if schema.Ref != "" {
		// Call top-level GetExample that uses the cache
		// to prevent reference cycles
		return gen.GetExample(schema.Ref)
	}
	// Return constant value if set
	if schema.Const != nil {
		return schema.Const, true
	}
	// Return first example if set
	if len(schema.Examples) > 0 {
		return schema.Examples[0], true
	}
	// Return "example" value from extras if set
	if example := getExtraKey(schema, "example"); example != nil {
		return example, true
	}
	// Return first enum value as enum example
	if len(schema.Enum) > 0 {
		return schema.Enum[0], true
	}

	// Schema type switch
	switch schema.Type {
	// Generate array element example and return a single-element array
	case TypeArray:
		return gen.generateArrayExample(loc, schema)
	// Generate field examples and return a map
	case TypeObject:
		return gen.generateObjectExample(loc, schema)
	// Return false as boolean example
	case TypeBoolean:
		return false, true
	// Return string examples
	case TypeString:
		switch schema.Format {
		case FormatDate:
			return "2006-01-02", true
		case FormatDateTime:
			return "2006-01-02T15:04:05Z", true
		case FormatTime:
			return "15:04:05Z", true
		default:
			return "string", true
		}
	// Return number examples
	case TypeNumber:
		switch schema.Format {
		case FormatFloat, FormatDouble:
			return "1000.123", true
		default:
			return "1", true
		}
	// Return number examples
	case TypeInteger:
		return "1", true
	}

	// Basic support for following "allOf" schemas:
	if example, ok := gen.generateAllOfExample(loc, schema); ok {
		return example, ok
	}
	// Basic support for following "anyOf" schemas:
	if example, ok := gen.generateAnyOfExample(loc, schema); ok {
		return example, ok
	}
	// Basic support for following "oneOf" schemas:
	if example, ok := gen.generateOneOfExample(loc, schema); ok {
		return example, ok
	}

	// Log error and return false
	slog.Error("empty example for schema", slog.String("location", loc))
	return nil, false
}

func getExtraKey(schema *jsonschema.Schema, key string) any {
	if schema == nil || schema.Extra == nil {
		return nil
	}
	return schema.Extra[key]
}

// generateArrayExample generates an example value for an array schema.
func (gen *ExampleGenerator) generateArrayExample(loc string, schema *jsonschema.Schema) ([]any, bool) {
	example, ok := gen.generateSchemaExample(loc+"/items", schema.Items)
	if !ok {
		slog.Error("generating example for array element", slog.String("location", loc))
		// Return nil array
		return nil, false
	}
	if schema.MinItems != nil {
		// Ensure value is repeated the minimum number of times
		return slices.Repeat([]any{example}, int(*schema.MinItems)), true
	}
	// Return example as an array
	return []any{example}, true
}

// generateObjectExample generates an example value for an object schema.
func (gen *ExampleGenerator) generateObjectExample(loc string, schema *jsonschema.Schema) (*orderedmap.OrderedMap, bool) {
	if schema == nil {
		return nil, false
	}
	// Use orderedmap for properties
	// This makes the JSON output deterministic and
	// easier to compare to the schema
	example := orderedmap.New()
	// Use to check if property is required
	required := toLookupMap(schema.Required)
	// Create examples for each property
	for propName, propSchema := range OrderedProperties(schema) {
		propLocation := loc + "/" + jsonpointer.Escape(propName)
		propExample, ok := gen.generateSchemaExample(propLocation, propSchema)
		if !ok {
			if required[propName] {
				slog.Error("no example for required property",
					slog.String("location", loc),
					slog.String("property", propName),
				)
			}
			continue
		}
		example.Set(propName, propExample)
	}

	// Add example for additional properties
	if addPropSchema := schema.AdditionalProperties; addPropSchema != nil {
		addPropName, addPropExample, ok := gen.generateAdditionalPropertiesExample(loc+"/additionalProperties", addPropSchema)
		if ok {
			example.Set(addPropName, addPropExample)
		}
	}

	return example, true
}

// generateAdditionalPropertiesExample generates an example value for an additionalProperties schema.
func (gen *ExampleGenerator) generateAdditionalPropertiesExample(loc string, addPropSchema *jsonschema.Schema) (addPropName string, addPropExample any, ok bool) {
	// Support for the "x-additionalPropertiesName" extension
	addPropName = cmp.Or(GetXAdditionalPropertiesName(addPropSchema), "additionalProp1")
	addPropExample, ok = gen.generateSchemaExample(loc, addPropSchema)
	if !ok {
		slog.Error("no example for additional property schema",
			slog.String("location", loc),
		)
	}
	return addPropName, addPropExample, ok
}

// Basic support for following "allOf" schemas
func (gen *ExampleGenerator) generateAllOfExample(loc string, schema *jsonschema.Schema) (any, bool) {
	index, allOfSchema := enterAllOf(loc, schema)
	if index < 0 {
		return nil, false
	}
	subLoc := loc + "/allOf/" + strconv.Itoa(index)
	example, ok := gen.generateSchemaExample(subLoc, allOfSchema)
	if !ok {
		slog.Error("empty example for allOf schema", slog.String("location", subLoc))
	}
	return example, ok
}

// Basic support for following "anyOf" schemas
func (gen *ExampleGenerator) generateAnyOfExample(loc string, schema *jsonschema.Schema) (any, bool) {
	index, anyOfSchema := enterAnyOf(loc, schema)
	if index < 0 {
		return nil, false
	}
	subLoc := loc + "/anyOf/" + strconv.Itoa(index)
	example, ok := gen.generateSchemaExample(subLoc, anyOfSchema)
	if !ok {
		slog.Error("empty example for anyOf schema", slog.String("location", subLoc))
	}
	return example, ok
}

// Basic support for following "oneOf" schemas
func (gen *ExampleGenerator) generateOneOfExample(loc string, schema *jsonschema.Schema) (any, bool) {
	index, oneOfSchema := enterOneOf(loc, schema)
	if index < 0 {
		return nil, false
	}
	subLoc := loc + "/oneOf/" + strconv.Itoa(index)
	example, ok := gen.generateSchemaExample(subLoc, oneOfSchema)
	if !ok {
		slog.Error("empty example for oneOf schema", slog.String("location", subLoc))
	}
	return example, ok
}

func toLookupMap[K comparable](slice []K) map[K]bool {
	out := make(map[K]bool, len(slice))
	for _, key := range slice {
		out[key] = true
	}
	return out
}

func enterAllOf(loc string, schema *jsonschema.Schema) (int, *jsonschema.Schema) {
	switch len(schema.AllOf) {
	// No allOf value
	case 0:
		return -1, nil
	// Enter single allOf value
	case 1:
		return 0, schema.AllOf[0]
	// Log error but still enter first allOf value
	default:
		// Cannot use the enterSubschemaList function because allOf is different
		slog.Warn("ignoring multiple allOf schemas for example",
			slog.String("location", loc+"/allOf"),
			slog.Int("length", len(schema.AllOf)),
		)
		return 0, schema.AllOf[0]
	}
}

func enterAnyOf(loc string, schema *jsonschema.Schema) (int, *jsonschema.Schema) {
	switch len(schema.AnyOf) {
	// No anyOf value
	case 0:
		return -1, nil
	// Enter single anyOf value
	case 1:
		return 0, schema.AnyOf[0]
	// Select one of the subschemas
	default:
		return enterSubschemaList(loc, "anyOf", schema.AnyOf)
	}
}

func enterOneOf(loc string, schema *jsonschema.Schema) (int, *jsonschema.Schema) {
	switch len(schema.OneOf) {
	// No oneOf value
	case 0:
		return -1, nil
	// Enter single oneOf value
	case 1:
		return 0, schema.OneOf[0]
	// Log error but still enter first oneOf value
	default:
		return enterSubschemaList(loc, "oneOf", schema.OneOf)
	}
}

func enterSubschemaList(loc string, key string, schemas []*jsonschema.Schema) (int, *jsonschema.Schema) {
	var (
		selectedIndex  = -1
		selectedSchema *jsonschema.Schema
	)
	for i, subSchema := range schemas {
		if subSchema != nil && subSchema.Type != TypeNull {
			// Stop early if subschema is not type=null
			selectedIndex = i
			selectedSchema = subSchema
			break
		} else if selectedSchema == nil {
			// Save non-nil subschemas as fallback
			selectedIndex = i
			selectedSchema = subSchema
		}
	}
	if selectedSchema != nil {
		slog.Info("selecting "+key+" schema for example",
			slog.String("location", loc+"/"+key+"/"+strconv.Itoa(selectedIndex)),
			slog.Int("length", len(schemas)),
		)
	}
	return selectedIndex, selectedSchema
}
