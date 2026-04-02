package schemagen_test

import (
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

	got, err := schemagen.GenerateSchemaForType(gen, tt.Type)
	if testutil.AssertErrorIf(t, tt.WantErr != "", err) && tt.WantErr != "" {
		assert.ErrorContains(t, err, tt.WantErr)
	}
	assert.Equal(t, tt.Schema, got)
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
	AThird string `json:"aThird,omitzero"`

	FieldWithOmitemptyOmitzero string `json:"fieldWithOmitemptyOmitzero,omitempty,omitzero"`

	// This field will not be included in the schema.
	IgnoredField int64 `json:"-"`

	// This field is named -.
	DashField int64 `json:"-,"`

	// This field is an array.
	ArrayField [3]float64 `json:"arrayField,omitzero"`

	// This field does not have a struct tag.
	NoStructTag string
}

func TestGenerateSchema(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		tt := TestCase{
			Type: reflect.TypeFor[TestStruct](),
			Schema: &jsonschema.Schema{
				Type:        "object",
				Description: "A struct type.",
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
					},
					"fieldWithOmitemptyOmitzero": {
						Type: "string",
					},
					"-": {
						Type:        "integer",
						Format:      "int64",
						Description: "This field is named -.",
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
				},
				PropertyOrder: []string{"cFirst", "bSecond", "aThird", "fieldWithOmitemptyOmitzero", "-", "arrayField", "NoStructTag"},
				Required:      []string{"cFirst", "-", "NoStructTag"},
			},
		}

		RunTestCase(t, tt)
	})
}
