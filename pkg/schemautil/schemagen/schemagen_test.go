package schemagen_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	"github.com/act3-ai/go-common/pkg/astutil"
	"github.com/act3-ai/go-common/pkg/schemautil"
	"github.com/act3-ai/go-common/pkg/schemautil/schemagen"
	"github.com/act3-ai/go-common/pkg/testutil"
)

type TestCase struct {
	Type    reflect.Type
	Schema  *jsonschema.Schema
	WantErr string
}

func RunTestCase(t *testing.T, tt TestCase) {
	t.Helper()

	info, err := astutil.LoadPackageInfo(t.Context(), []string{"./..."}, func(cfg *packages.Config) {
		cfg.Tests = true
	})
	require.NoError(t, err)

	gen := schemagen.NewGenerator()
	gen.PackageInfo = info
	gen.SetXOrder = true

	got, err := gen.GenerateSchemaForType(tt.Type)
	if testutil.AssertErrorIf(t, tt.WantErr != "", err) && tt.WantErr != "" {
		assert.Equal(t, tt.WantErr, err.Error())
	}
	assert.Equal(t, tt.Schema, got)

	data, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)
	t.Log(string(data))
}

// A struct type to be embedded.
type EmbeddedStruct struct {
	// An embedded field.
	EmbeddedField []string `json:"embeddedField"`
}

// A struct type.
//
//directive:command args args args
type TestStruct struct {
	// A string field.
	CFirst string `json:"cFirst"`

	// Another string field.
	//
	//directive:command args args args
	BSecond string `json:"bSecond,omitempty"`

	// This is yet another string field.
	//
	//jsonschema:set "$comment" "This is set by a directive"
	AThird string `json:"aThird,omitzero"`

	FieldWithOmitemptyOmitzero string `json:"fieldWithOmitemptyOmitzero,omitempty,omitzero"`

	NoComments string `json:"noComments,omitzero"`

	//tool:name args args args
	OnlyDirectiveComments string `json:"OnlyDirectiveComments,omitzero"`

	// This field will not be included in the schema.
	IgnoredField int64 `json:"-"`

	// This field is named -.
	//
	//jsonschema:set maximum 5
	DashField int64 `json:"-,"`

	// This field is an array.
	ArrayField [3]float64 `json:"arrayField,omitzero"`

	// This field does not have a struct tag.
	NoStructTag string

	// This is an embedded field.
	EmbeddedStruct

	// This is a required field with omitempty.
	//
	//jsonschema:required
	RequiredWithOmitempty string `json:"requiredWithOmitempty,omitempty"`

	// This is a required field with omitzero.
	//
	//jsonschema:required true
	RequiredWithOmitzero string `json:"requiredWithOmitzero,omitzero"`

	// This field has been manually set to be optional.
	//
	//jsonschema:required false
	ManuallyNotRequired string `json:"ManuallyNotRequired"`

	// This field is the type any.
	AnyField any

	// This field is a json.RawMessage.
	RawMessageField json.RawMessage
}

func TestGenerateSchema(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		tt := TestCase{
			Type: reflect.TypeFor[TestStruct](),
			Schema: &jsonschema.Schema{
				Ref: "#/$defs/schemagen_test.TestStruct",
				Defs: map[string]*jsonschema.Schema{
					"schemagen_test.EmbeddedStruct": {
						Type:        "object",
						Description: "A struct type to be embedded.",
						Properties: map[string]*jsonschema.Schema{
							"embeddedField": {
								Type:        "array",
								Description: "An embedded field.",
								Items: &jsonschema.Schema{
									Type: "string",
								},
								Extra: map[string]any{
									"x-order": 1,
								},
							},
						},
						PropertyOrder: []string{
							"embeddedField",
						},
						Required: []string{
							"embeddedField",
						},
						AdditionalProperties: schemautil.FalseSchema(),
					},
					"schemagen_test.TestStruct": {
						Description: "A struct type.",
						AllOf: []*jsonschema.Schema{
							{
								Description: "This is an embedded field.",
								AllOf: []*jsonschema.Schema{
									{
										Ref: "#/$defs/schemagen_test.EmbeddedStruct",
									},
								},
							},
							{
								Type: "object",
								Properties: map[string]*jsonschema.Schema{
									"cFirst": {
										Type:        "string",
										Description: "A string field.",
										Extra: map[string]any{
											"x-order": 1,
										},
									},
									"bSecond": {
										Type:        "string",
										Description: "Another string field.",
										Extra: map[string]any{
											"x-order": 2,
										},
									},
									"aThird": {
										Type:        "string",
										Description: "This is yet another string field.",
										Comment:     "This is set by a directive",
										Extra: map[string]any{
											"x-order": 3,
										},
									},
									"fieldWithOmitemptyOmitzero": {
										Type: "string",
										Extra: map[string]any{
											"x-order": 4,
										},
									},
									"noComments": {
										Type: "string",
										Extra: map[string]any{
											"x-order": 5,
										},
									},
									"OnlyDirectiveComments": {
										Type: "string",
										Extra: map[string]any{
											"x-order": 6,
										},
									},
									"-": {
										Type:        "integer",
										Format:      "int64",
										Description: "This field is named -.",
										Maximum:     new(5.),
										Extra: map[string]any{
											"x-order": 7,
										},
									},
									"arrayField": {
										Type:        "array",
										Description: "This field is an array.",
										Items: &jsonschema.Schema{
											Type:   "number",
											Format: "double",
										},
										MinItems: new(3),
										MaxItems: new(3),
										Extra: map[string]any{
											"x-order": 8,
										},
									},
									"NoStructTag": {
										Type:        "string",
										Description: "This field does not have a struct tag.",
										Extra: map[string]any{
											"x-order": 9,
										},
									},
									"requiredWithOmitempty": {
										Type:        "string",
										Description: "This is a required field with omitempty.",
										Extra: map[string]any{
											"x-order": 10,
										},
									},
									"requiredWithOmitzero": {
										Type:        "string",
										Description: "This is a required field with omitzero.",
										Extra: map[string]any{
											"x-order": 11,
										},
									},
									"ManuallyNotRequired": {
										Type:        "string",
										Description: "This field has been manually set to be optional.",
										Extra: map[string]any{
											"x-order": 12,
										},
									},
									"AnyField": {
										Description: "This field is the type any.",
										Extra: map[string]any{
											"x-order": 13,
										},
									},
									"RawMessageField": {
										Description: "This field is a json.RawMessage.",
										Extra: map[string]any{
											"x-order": 14,
										},
									},
								},
								PropertyOrder: []string{
									"cFirst",
									"bSecond",
									"aThird",
									"fieldWithOmitemptyOmitzero",
									"noComments",
									"OnlyDirectiveComments",
									"-",
									"arrayField",
									"NoStructTag",
									"requiredWithOmitempty",
									"requiredWithOmitzero",
									"ManuallyNotRequired",
									"AnyField",
									"RawMessageField",
								},
								Required: []string{
									"cFirst",
									"-",
									"NoStructTag",
									"requiredWithOmitempty",
									"requiredWithOmitzero",
									"AnyField",
									"RawMessageField",
								},
								AdditionalProperties: schemautil.FalseSchema(),
							},
						},
					},
				},
			},
		}

		RunTestCase(t, tt)
	})

	t.Run("string", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type: reflect.TypeFor[string](),
			Schema: &jsonschema.Schema{
				Type: "string",
			},
			WantErr: "",
		})
	})
	t.Run("bool", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type: reflect.TypeFor[bool](),
			Schema: &jsonschema.Schema{
				Type: "boolean",
			},
			WantErr: "",
		})
	})
	t.Run("uint", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type: reflect.TypeFor[uint](),
			Schema: &jsonschema.Schema{
				Type:   "integer",
				Format: "uint32",
			},
			WantErr: "",
		})
	})
	t.Run("[3]float64", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type: reflect.TypeFor[[3]float64](),
			Schema: &jsonschema.Schema{
				Type: "array",
				Items: &jsonschema.Schema{
					Type:   "number",
					Format: "double",
				},
				MinItems: new(3),
				MaxItems: new(3),
			},
			WantErr: "",
		})
	})
	t.Run("[4]int", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type: reflect.TypeFor[[4]int](),
			Schema: &jsonschema.Schema{
				Type: "array",
				Items: &jsonschema.Schema{
					Type:   "integer",
					Format: "int32",
				},
				MinItems: new(4),
				MaxItems: new(4),
			},
			WantErr: "",
		})
	})
	t.Run("complex64", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type:    reflect.TypeFor[complex64](),
			Schema:  nil,
			WantErr: "generating schema for type complex64: unsupported type",
		})
	})
	t.Run("complex128", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type:    reflect.TypeFor[complex128](),
			Schema:  nil,
			WantErr: "generating schema for type complex128: unsupported type",
		})
	})
	t.Run("chan int", func(t *testing.T) {
		RunTestCase(t, TestCase{
			Type:    reflect.TypeFor[chan int](),
			Schema:  nil,
			WantErr: "generating schema for type chan int: unsupported type",
		})
	})
}
