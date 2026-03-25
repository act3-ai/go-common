package schemautil

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// Known OpenAPI and JSON Schema extensions.
//
// From OpenAPI Initiative: https://spec.openapis.org/registry/index.html
// From Redocly: https://redocly.com/docs/realm/content/api-docs/openapi-extensions
const (
	// Display a field name for an additionalProperties description.
	XAdditionalPropertiesName = "x-additionalPropertiesName"

	// Add visible badges as indicators to API operations.
	XBadges = "x-badges"

	// Provide the code sample to display for an operation.
	XCodeSamples = "x-codeSamples"

	// Readable labels for enum values.
	XEnumDescriptions = "x-enumDescriptions"

	// Promote or exclude description files, operations, or tags in search results for specified keywords.
	XKeywords = "x-keywords"

	// Add custom metadata at the top of the info section.
	XMetadata = "x-metadata"

	// Specify order of properties in an object schema for documentation purposes only.
	XOrder = "x-order"

	// Add individual schemas to navigation sections alongside operations.
	XTags = "x-tags"

	// Add short summary of the response.
	XSummary = "x-summary"
)

// GetXAdditionalPropertiesName returns the value of the "x-additionalPropertiesName" extension or the empty string.
func GetXAdditionalPropertiesName(addPropSchema *jsonschema.Schema) string {
	return GetExtensionString(addPropSchema, XAdditionalPropertiesName)
}

// GetExtensionString checks the schema for the extension and returns its value as a string if it is present.
// If the extension is not present, or could not be cast to a string, the empty string is returned.
func GetExtensionString(schema *jsonschema.Schema, extension string) string {
	if schema == nil || schema.Extra == nil {
		return ""
	}

	// Check extras map for extension value
	extValue, ok := schema.Extra[extension]
	if !ok {
		return ""
	}

	// Cast to string
	extValueStr, ok := extValue.(string)
	if !ok {
		return ""
	}

	return extValueStr
}

// GetExtensionInt checks the schema for the extension and returns its value as an int if it is present.
// If the extension is not present, or could not be cast to an int, the boolean value will be false.
func GetExtensionInt(schema *jsonschema.Schema, extension string) (int, bool) {
	if schema == nil || schema.Extra == nil {
		return 0, false
	}

	// Check extras map for extension value
	extValue, ok := schema.Extra[extension]
	if !ok {
		return 0, false
	}

	// Convert to int
	return convertInt(extValue)
}

// GetPropertyOrder produces a canonical order of the schema's properties based on the
// "x-order" extension, with alphabetical sorting as a fallback.
func GetPropertyOrder(schema *jsonschema.Schema) []string {
	// Store x-order attributes for later use in sorting
	order := map[string]int{}
	// List of property names
	props := make([]string, 0, len(schema.Properties))
	for propName, propSchema := range schema.Properties {
		// Check for "x-order" extension
		propOrder, ok := GetExtensionInt(propSchema, XOrder)
		if ok {
			order[propName] = propOrder
		}
		props = append(props, propName)
	}
	if len(props) == 0 {
		return nil
	}

	// Sort the property names by x-order attribute
	slices.SortFunc(props, func(first, second string) int {
		cmp := order[first] - order[second]
		if cmp == 0 {
			// Default to alphabetical order
			return strings.Compare(first, second)
		}
		return cmp
	})
	return props
}

// SetPropertyOrder sets the property order for a schema.
func SetPropertyOrder(schema *jsonschema.Schema) {
	// Set property order from x-order extension
	schema.PropertyOrder = GetPropertyOrder(schema)
}

func convertInt(v any) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		// Parse as int64
		parsedInt, err := v.Int64()
		if err == nil {
			return int(parsedInt), true
		}
		// Parse as float64 and convert to integer
		parsedFloat, err := v.Float64()
		if err == nil {
			return int(parsedFloat), true
		}
	}
	return 0, false
}
