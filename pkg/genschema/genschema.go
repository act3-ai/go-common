/*
Package genschema generates JSON Schema definitions for Go types.

The JSON Schema generation uses [invopop/jsonschema], which is based on type reflection.

The schema definitions are intended to be embedded in a Go CLI binary to be "generated" on the user's system. Below is an example of how to generate the schema use the go:generate directive:

	package gen

	//go:generate go run internal/gen/main.go

And the main.go file called by "//go:generate go run internal/gen/main.go"

	//go:build ignore

	package main

	import (
		"fmt"
		"log"

		"git.act3-ace.com/ace/go-common/pkg/genschema"
		"git.act3-ace.com/devsecops/act3-pt/pkg/apis/pt.act3-ace.io/v1alpha3"
	)

	func main() {
	 	// Generate JSON Schema definitions
	 	if err := genschema.GenJSONSchema(
	 		"cmd/act3-pt/schemas",
	 		[]any{&v1alpha3.Project{}, &v1alpha3.Template{}, &v1alpha3.Configuration{}},
	 		"pt.act3-ace.io/v1alpha3",
	 		"git.act3-ace.com/devsecops/act3-pt",
	 	); err != nil {
	 		log.Fatal(fmt.Errorf("JSON Schema generation failed: %w", err))
	 	}
	}

And finally, embedding the JSON Schema definitions and adding a "genschema" command:

	//go:embed schemas/*
	var schemaDefs embed.FS

	associations := []SchemaAssociation{
		{
			Definition: "schemas/project-schema.json",
			FileMatch:  []string{".act3-pt.yaml"},
		},
		{
			Definition: "schemas/template-shema.json",
			FileMatch:  []string{".act3-template.yaml"},
		},
	}

	NewGenschemaCmd(schemaDefs, associations)
*/
package genschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"
)

// GenJSONSchema generates JSON Schema definitions for internal Go types
//
// - schemas is a list of types (schema) to a generate schema for.
// - baseSchemaID is the base name for the schema definitions. Use "apiVersion" values for KRM file schemas.
// - moduleName is used to add Go comments to the schema as descriptions, pass an empty string to disable this.
//
//	GenJSONSchema("schemas", []any{&v1alpha3.Project{}, &v1alpha3.Template{}}, "pt.act3-ace.io/v1alpha3", "git.act3-ace.com/devsecops/act3-pt")
func GenJSONSchema(schemaDir string, schemas []any, baseSchemaID string, moduleName string) error {
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	/*
		JSON Schema Generator Setup

		AddGoComments: enables setting descriptions from Go comments
		SetBaseSchemaID: changes the $id field of the schema to start with this string
								rather than the module path
	*/

	r := new(jsonschema.Reflector)

	if moduleName != "" {
		// WARNING: because of the "./" argument, this only works when running on the source files
		// 	This can cause errors when running in an executable since it will try to parse any .go files
		// 	on the system. This limitation is why we generate the schema at build time and embed the
		// 	schema files into the executable.
		err := r.AddGoComments(moduleName, "./")
		if err != nil {
			return fmt.Errorf("could not add comments to schema generator: %w", err)
		}
	}

	// JSON Schema convention is to include "https://" for URLs
	if !strings.HasPrefix(baseSchemaID, "https://") || !strings.HasPrefix(baseSchemaID, "http://") {
		baseSchemaID = "https://" + baseSchemaID
	}
	r.SetBaseSchemaID(baseSchemaID)

	// Iterate over each schema that needs generated
	for _, schema := range schemas {
		// Create the JSON Schema
		_, err := generateSchema(r, schemaDir, schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateSchema(r *jsonschema.Reflector, dir string, schemaType any) (string, error) {
	// Create the JSON Schema
	schema := r.Reflect(schemaType)

	bts, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to create jsonschema: %w", err)
	}

	// Write JSON Schema definition to a file
	// Derive file name from "schema.ID", format is Go type name in lowercase
	schemaFile := filepath.Join(dir, filepath.Base(schema.ID.Base().String())+"-schema.json")
	if err := os.WriteFile(schemaFile, bts, 0o666); err != nil {
		return schemaFile, fmt.Errorf("failed to write jsonschema file: %w", err)
	}

	return schemaFile, nil
}
