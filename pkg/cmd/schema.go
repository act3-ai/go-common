package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"
)

// Schema represents a JSON Schema definition to generate
//
// Example:
//
//	schemas := []Schema{
//		{
//			Type:      &v1alpha3.Project{},
//			FileMatch: []string{".act3-pt.yaml"},
//		},
//		{
//			Type:      &v1alpha3.Template{},
//			FileMatch: []string{".act3-template.yaml"},
//		},
//	}
type Schema struct {
	Type      any      // The type representing the schema
	FileMatch []string // List of filenames to validate with the schema
}

// NewSchemaCmd creates a command to generate the internal schema definitions in JSONSchema
// schemaMap is a map of types (schema) to a list of patterns for files that should match the schema
//
// Example:
//
//	schemas := []Schema{
//		{
//			Type:      &v1alpha3.Project{},
//			FileMatch: []string{".act3-pt.yaml"},
//		},
//		{
//			Type:      &v1alpha3.Template{},
//			FileMatch: []string{".act3-template.yaml"},
//		},
//	}
//
//	NewSchemaCmd("git.act3-ace.com/devsecops/act3-pt", "pt.act3-ace.io/v1alpha3", schemas)
func NewSchemaCmd(module string, baseSchemaID string, schemas []Schema) *cobra.Command {
	var schemaCmd = &cobra.Command{
		Use:   "genschema <schema location>",
		Short: "Outputs configuration file validators",
		Long: `Outputs schema definitions for configuration files in JSON Schema format.
Provides instructions for adding the schema definitions to VS Code to validate configuration files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaDir, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("could not evaluate output directory: %w", err)
			}

			if err := os.MkdirAll(schemaDir, 0o755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			/*
				JSON Schema Generator Setup

				AddGoComments: enables setting descriptions from Go comments
				SetBaseSchemaID: changes the $id field of the schema to start with this string
										rather than the module path

			*/

			r := new(jsonschema.Reflector)

			err = r.AddGoComments(module, "./")
			if err != nil {
				return fmt.Errorf("could not add comments to schema generator: %w", err)
			}

			r.SetBaseSchemaID(baseSchemaID)

			/*
				Iterate over each schema that needs generated
			*/

			yamlSettings := vsCodeYAMLSchemaSettings{}
			jsonSettings := vsCodeJSONSchemaSettings{}

			for _, schema := range schemas {
				// Create the JSON Schema
				schemaFile, err := generateSchema(r, schemaDir, schema.Type)
				if err != nil {
					return err
				}

				// Build the VS Code settings to associate the schema with files
				newYAML, newJSON := generateVSCodeSettings(schemaFile, schema.FileMatch)

				// Add the settings to the global settings
				yamlSettings.add(newYAML)
				jsonSettings.add(newJSON)
			}

			yamlout, err := yamlSettings.marshal()
			if err != nil {
				return err
			}
			if len(yamlout) > 0 {
				cmd.Println("Add the following to VS Code's settings.json file to enable YAML file validation:\n\n" + yamlout + "\n")
			}

			jsonout, err := jsonSettings.marshal()
			if err != nil {
				return err
			}
			if len(jsonout) > 0 {
				cmd.Println("Add the following to VS Code's settings.json file to enable JSON file validation:\n\n" + jsonout + "\n")
			}

			return nil
		},
	}

	return schemaCmd
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

func generateVSCodeSettings(schemaFile string, fileMatches []string) (yamlRule vsCodeYAMLSchemaSettings, jsonRule vsCodeJSONSchemaSettings) {
	// VS Code requires local file paths begin with "file://"
	schemaFileURI := "file://" + schemaFile

	yamlFiles := []string{}
	jsonFiles := []string{}

	// Process file matches to output settings to add to VS Code
	for _, pattern := range fileMatches {
		switch filepath.Ext(pattern) {
		case ".yaml", ".yml":
			// Add entry to the YAML schemas map
			yamlFiles = append(yamlFiles, pattern)
		case ".json":
			// Add to list of file matches
			jsonFiles = append(jsonFiles, pattern)
		}
	}

	// Only add the YAML setting if there were YAML files given
	switch length := len(yamlFiles); {
	case length == 1:
		// Add as a string for single file association
		yamlRule = vsCodeYAMLSchemaSettings{
			schemaFileURI: yamlFiles[0],
		}
	case length > 1:
		// Add as a list for multiple file associations
		yamlRule = vsCodeYAMLSchemaSettings{
			schemaFileURI: yamlFiles,
		}
	}

	// Only add the JSON setting if there were JSON files given
	if len(jsonFiles) > 0 {
		jsonRule = vsCodeJSONSchemaSettings{
			{
				URL:       schemaFileURI,
				FileMatch: jsonFiles,
			},
		}
	}

	return yamlRule, jsonRule
}

/*
Example VS Code YAML schemas setting:

	"yaml.schemas": {
		"file:///Users/username/.config/act3/pt/schema/template-schema.json": ".act3-template.yaml",
		"https://goreleaser.com/static/schema.json": ".goreleaser.yaml",
	},
*/
type vsCodeYAMLSchemaSettings map[string]any

func (s *vsCodeYAMLSchemaSettings) add(newSettings vsCodeYAMLSchemaSettings) {
	maps.Copy(*s, newSettings)
}

func (s vsCodeYAMLSchemaSettings) marshal() (string, error) {
	if len(s) == 0 {
		return "", nil
	}

	yamlout, err := json.MarshalIndent(s, "  ", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to create settings JSON: %w", err)
	}

	return "  \"yaml.schemas\": " + string(yamlout), nil
}

/*
Example VS Code JSON schemas setting:

	"json.schemas": [
		{
			"fileMatch": [
				"validatethis.json"
			],
			"url": "file:///abs/path/to/schema.json"
		}
	]
*/
type vsCodeJSONSchemaSettings []vsCodeJSONSchemaSetting

type vsCodeJSONSchemaSetting struct {
	FileMatch []string `json:"fileMatch"`
	URL       string   `json:"url"`
}

func (s *vsCodeJSONSchemaSettings) add(newSettings vsCodeJSONSchemaSettings) {
	*s = append(*s, newSettings...)
}

func (s vsCodeJSONSchemaSettings) marshal() (string, error) {
	if len(s) == 0 {
		return "", nil
	}

	jsonout, err := json.MarshalIndent(s, "  ", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to create settings JSON: %w", err)
	}

	return "  \"json.schemas\": " + string(jsonout), nil
}
