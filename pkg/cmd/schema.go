package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewSchemaCmd is a command to generate the internal schema definitions in JSONSchema
func NewSchemaCmd(schemaDefinitions fs.FS, fileAssociations map[string]string) *cobra.Command {
	var schemaCmd = &cobra.Command{
		Use:    "genschema <docs location>",
		Short:  "Generate JSONSchema schema definitions",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaDir := args[0]

			if schemaDefinitions == nil {
				return nil
			}

			err := fs.WalkDir(schemaDefinitions, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				src, err := schemaDefinitions.Open(path)
				if err != nil {
					return fmt.Errorf("could not open schema definition %q: %w", path, err)
				}

				dst, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("could not create file %q: %w", path, err)
				}

				if _, err = io.Copy(dst, src); err != nil {
					return fmt.Errorf("could not copy content to %q: %w", dst.Name(), err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("could not generate JSONSchema schema definitions: %w", err)
			}

			absSchemaDir, err := filepath.Abs(schemaDir)
			if err != nil {
				return fmt.Errorf("cannot generate VS Code settings: %w", err)
			}

			jsonSchemas := []string{}
			yamlSchemas := []string{}

			for fileMatch, schema := range fileAssociations {
				schemaPath := "file://" + filepath.Join(absSchemaDir, schema)
				jsonSchemas = append(jsonSchemas, fmt.Sprintf(`
	{
		"fileMatch": [
			"%s"
		],
		"url": "%s"
	}`, fileMatch, schemaPath))

				yamlSchemas = append(yamlSchemas, fmt.Sprintf(`
"%s": "%s"`, schemaPath, fileMatch))
			}

			cmd.Println("To use the schemas for validation in VS Code, add the following to VS Code's settings.json file:")

			cmd.Printf(`"json.schemas": [
` + strings.Join(jsonSchemas, ",") + `],
`)
			cmd.Printf(`"yaml.schemas": {
` + strings.Join(yamlSchemas, ",") + "}\n")

			return nil
		},
	}
	return schemaCmd
}
