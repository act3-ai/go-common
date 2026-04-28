// package invopopadapter converts schemas between [github.com/invopop/jsonschema] and [github.com/google/jsonschema-go/jsonschema].
package invopopadapter

import (
	"encoding/json"
	"fmt"
	"math"

	gschema "github.com/google/jsonschema-go/jsonschema"
	ischema "github.com/invopop/jsonschema"
	orderedmap "github.com/pb33f/ordered-map/v2"

	"github.com/act3-ai/go-common/pkg/schemautil"
)

// ToGoogleJSONSchema converts a schema from [github.com/invopop/jsonschema] to [github.com/google/jsonschema-go/jsonschema].
//
// This function panics on any error.
func ToGoogleJSONSchema(in *ischema.Schema) *gschema.Schema {
	if in == nil {
		return nil
	}
	// Special cases for the "true" or "false" schema
	switch in {
	case ischema.TrueSchema:
		return schemautil.TrueSchema()
	case ischema.FalseSchema:
		return schemautil.FalseSchema()
	}
	out := &gschema.Schema{
		ID:      string(in.ID),
		Schema:  in.Version,
		Ref:     in.Ref,
		Comment: in.Comments,
		Defs:    convertSchemaMap(in.Definitions),
		// Definitions: N/A,
		// DependencySchemas: N/A,
		// DependencyStrings: N/A,
		Anchor: in.Anchor,
		// DynamicAnchor: N/A,
		DynamicRef: in.DynamicRef,
		// Vocabulary: N/A,
		Title:       in.Title,
		Description: in.Description,
		Default:     must(jsonMarshalIfNotNil(in.Default)),
		Deprecated:  in.Deprecated,
		ReadOnly:    in.ReadOnly,
		WriteOnly:   in.WriteOnly,
		Examples:    in.Examples,
		Type:        in.Type,
		// Types: N/A,
		Enum:             in.Enum,
		Const:            ptrIfNotNil(in.Const),
		MultipleOf:       must(convertNumberToFloat64(in.MultipleOf)),
		Minimum:          must(convertNumberToFloat64(in.Minimum)),
		Maximum:          must(convertNumberToFloat64(in.Maximum)),
		ExclusiveMinimum: must(convertNumberToFloat64(in.ExclusiveMinimum)),
		ExclusiveMaximum: must(convertNumberToFloat64(in.ExclusiveMaximum)),
		MinLength:        must(convertUint64ToInt(in.MinLength)),
		MaxLength:        must(convertUint64ToInt(in.MaxLength)),
		Pattern:          in.Pattern,
		PrefixItems:      convertSchemaSlice(in.PrefixItems),
		Items:            ToGoogleJSONSchema(in.Items),
		// ItemsArray: N/A,
		MinItems: must(convertUint64ToInt(in.MinItems)),
		MaxItems: must(convertUint64ToInt(in.MaxItems)),
		// AdditionalItems: N/A,
		UniqueItems: in.UniqueItems,
		Contains:    ToGoogleJSONSchema(in.Contains),
		MinContains: must(convertUint64ToInt(in.MinContains)),
		MaxContains: must(convertUint64ToInt(in.MaxContains)),
		// UnevaluatedItems: N/A,
		MinProperties:        must(convertUint64ToInt(in.MinProperties)),
		MaxProperties:        must(convertUint64ToInt(in.MaxProperties)),
		Required:             in.Required,
		DependentRequired:    in.DependentRequired,
		Properties:           convertSchemaOrderedMap(in.Properties),
		PatternProperties:    convertSchemaMap(in.PatternProperties),
		AdditionalProperties: ToGoogleJSONSchema(in.AdditionalProperties),
		PropertyNames:        ToGoogleJSONSchema(in.PropertyNames),
		// UnevaluatedProperties: N/A,
		AllOf:            convertSchemaSlice(in.AllOf),
		AnyOf:            convertSchemaSlice(in.AnyOf),
		OneOf:            convertSchemaSlice(in.OneOf),
		Not:              ToGoogleJSONSchema(in.Not),
		If:               ToGoogleJSONSchema(in.If),
		Then:             ToGoogleJSONSchema(in.Then),
		Else:             ToGoogleJSONSchema(in.Else),
		DependentSchemas: convertSchemaMap(in.DependentSchemas),
		ContentEncoding:  in.ContentEncoding,
		ContentMediaType: in.ContentMediaType,
		ContentSchema:    ToGoogleJSONSchema(in.ContentSchema),
		Format:           in.Format,
		Extra:            in.Extras,
		PropertyOrder:    keysFromOrderedMap(in.Properties),
	}
	return out
}

func convertSchemaSlice(in []*ischema.Schema) []*gschema.Schema {
	if in == nil {
		return nil
	}
	out := make([]*gschema.Schema, 0, len(in))
	for _, schema := range in {
		out = append(out, ToGoogleJSONSchema(schema))
	}
	return out
}

func convertSchemaMap(in map[string]*ischema.Schema) map[string]*gschema.Schema {
	if in == nil {
		return nil
	}
	out := make(map[string]*gschema.Schema, len(in))
	for name, schema := range in {
		out[name] = ToGoogleJSONSchema(schema)
	}
	return out
}

func convertSchemaOrderedMap(in *orderedmap.OrderedMap[string, *ischema.Schema]) map[string]*gschema.Schema {
	if in == nil {
		return nil
	}
	out := make(map[string]*gschema.Schema, in.Len())
	for pair := in.Oldest(); pair != nil; pair = pair.Next() {
		out[pair.Key] = ToGoogleJSONSchema(pair.Value)
	}
	return out
}

func keysFromOrderedMap(in *orderedmap.OrderedMap[string, *ischema.Schema]) []string {
	if in == nil {
		return nil
	}
	out := make([]string, 0, in.Len())
	for pair := in.Oldest(); pair != nil; pair = pair.Next() {
		out = append(out, pair.Key)
	}
	return out
}

// jsonMarshalIfNotNil calls json.Marshal on the value if it is not nil.
func jsonMarshalIfNotNil(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// ptrIfNotNil returns a pointer to the value if it is non-nil.
func ptrIfNotNil(v any) *any {
	if v, ok := v.(*any); ok {
		// Always return pointer values as-is
		return v
	}
	if v == nil {
		// Return nil pointer if v is nil interface
		return nil
	}
	// Return pointer to non-nil interface
	return &v
}

// must panics if err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// convertNumberToInt converts a json.Number to a *float64.
func convertNumberToFloat64(n json.Number) (*float64, error) {
	if len(n) == 0 {
		return nil, nil
	}
	float64Value, err := n.Float64()
	if err != nil {
		return nil, err
	}
	if float64Value > math.MaxFloat64 {
		return nil, fmt.Errorf("overflow converting from json.Number to float64: %s > %f", n.String(), math.MaxFloat64)
	}
	return &float64Value, nil
}

// // convertNumberToInt converts a json.Number to a *int.
// func convertNumberToInt(n json.Number) (*int, error) {
// 	if len(n) == 0 {
// 		return nil, nil
// 	}
// 	int64Value, err := n.Int64()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if int64Value > math.MaxInt {
// 		return nil, fmt.Errorf("overflow converting from json.Number to int: %s > %d", n.String(), math.MaxInt)
// 	}
// 	intValue := int(int64Value)
// 	return &intValue, nil
// }

// convertUint64ToInt converts a *uint64 to a *int.
func convertUint64ToInt(v *uint64) (*int, error) {
	if v == nil {
		return nil, nil
	}
	uint64Value := *v
	if uint64Value > math.MaxInt {
		return nil, fmt.Errorf("overflow converting from uint64 to int: %v > %d", uint64Value, math.MaxInt)
	}
	intValue := int(uint64Value)
	return &intValue, nil
}
