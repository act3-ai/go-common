package schemagen_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	"github.com/act3-ai/go-common/pkg/astutil"
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

	gen := schemagen.NewGenerator().WithPackageInfo(info)

	got, err := gen.GenerateSchemaForType(tt.Type)
	if testutil.AssertErrorIf(t, tt.WantErr != "", err) && tt.WantErr != "" {
		assert.ErrorContains(t, err, tt.WantErr)
	}
	assert.Equal(t, tt.Schema, got)

	gotDefs := gen.Definitions()
	data, err := json.MarshalIndent(gotDefs, "", "  ")
	require.NoError(t, err)

	t.Fail()

	fmt.Println(string(data))
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
}

func TestGenerateSchema(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		tt := TestCase{
			Type: reflect.TypeFor[TestStruct](),
			Schema: &jsonschema.Schema{
				Description: "A struct type.",
				AllOf: []*jsonschema.Schema{
					{
						Description: "This is an embedded field.",
						Ref:         "#/$defs/github.com/act3-ai/go-common/pkg/schemautil/schemagen_test.EmbeddedStruct",
					},
					{
						Type: "object",
						Properties: map[string]*jsonschema.Schema{
							"cFirst": {
								Type:        "string",
								Description: "A string field.",
							},
							"bSecond": {
								Type:        "string",
								Description: "Another string field.",
							},
							"aThird": {
								Type:        "string",
								Description: "This is yet another string field.",
								Comment:     "This is set by a directive",
							},
							"fieldWithOmitemptyOmitzero": {
								Type: "string",
							},
							"noComments": {
								Type: "string",
							},
							"OnlyDirectiveComments": {
								Type: "string",
							},
							"-": {
								Type:        "integer",
								Format:      "int64",
								Description: "This field is named -.",
								Maximum:     new(5.),
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
							},
							"NoStructTag": {
								Type:        "string",
								Description: "This field does not have a struct tag.",
							},
							"requiredWithOmitempty": {
								Type:        "string",
								Description: "This is a required field with omitempty.",
							},
							"requiredWithOmitzero": {
								Type:        "string",
								Description: "This is a required field with omitzero.",
							},
							"ManuallyNotRequired": {
								Type:        "string",
								Description: "This field has been manually set to be optional.",
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
						},
						Required: []string{
							"cFirst",
							"-",
							"NoStructTag",
							"requiredWithOmitempty",
							"requiredWithOmitzero",
						},
					},
				},
			},
		}

		RunTestCase(t, tt)
	})
}
