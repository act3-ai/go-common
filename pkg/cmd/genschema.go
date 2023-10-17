package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// SchemaAssociation associates a JSON Schema definition to the files it validates
//
// Example:
//
//	associations := []SchemaAssociation{
//		{
//			Definition: "project-schema.json",
//			FileMatch:  []string{".act3-pt.yaml"},
//		},
//		{
//			Definition: "template-shema.json",
//			FileMatch:  []string{".act3-template.yaml"},
//		},
//	}
type SchemaAssociation struct {
	Definition string   // Path to the schema definition
	FileMatch  []string // List of filenames to validate with the schema
}

// NewGenschemaCmd creates the genschema command, which generates JSON Schema definitions.
//
// The JSON Schema definitions must be made available in the schemaDefs fs.FS by embedding them. The [go-common/pkg/genschema] package provides the functionality to generate the JSON Schema definitions from Go types at build time.
//
// The associations list is used to create a snippet of VS Code settings to enable YAML/JSON file validation using the generated schema definitions.
//
// Example:
//
//	//go:embed schemas/*
//	var schemaDefs embed.FS
//
//	associations := []SchemaAssociation{
//		{
//			Definition: "schemas/project-schema.json",
//			FileMatch:  []string{".act3-pt.yaml"},
//		},
//		{
//			Definition: "schemas/template-shema.json",
//			FileMatch:  []string{".act3-template.yaml"},
//		},
//	}
//
//	NewGenschemaCmd(schemaDefs, associations)
//
// [go-common/pkg/genschema]: https://git.act3-ace.com/ace/go-common/-/tree/main/pkg/genschema
func NewGenschemaCmd(schemaDefs fs.FS, associations []SchemaAssociation) *cobra.Command {
	var schemaCmd = &cobra.Command{
		Use:   "genschema <schema location>",
		Short: "Outputs configuration file validators",
		Long: `Outputs schema definitions for configuration files in JSON Schema format.
Provides instructions for adding the schema definitions to VS Code to validate configuration files.`,
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaDir, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("could not evaluate output directory: %w", err)
			}

			if err := os.MkdirAll(schemaDir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			/*
				Iterate over each schema that needs generated
			*/

			yamlSettings := vsCodeYAMLSchemaSettings{}
			jsonSettings := vsCodeJSONSchemaSettings{}

			if err = fs.WalkDir(schemaDefs, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				// Create file from fs.FS to schemaDir
				if err = copyFile(schemaDefs, schemaDir, path); err != nil {
					return fmt.Errorf("could not create schema definition %q: %w", path, err)
				}

				schemaFile := filepath.Join(schemaDir, path)

				for _, assoc := range associations {
					if path == assoc.Definition {
						// Build the VS Code settings to associate the schema with files
						newYAML, newJSON := generateVSCodeSettings(schemaFile, assoc.FileMatch)

						// Add the settings to the global settings
						yamlSettings.add(newYAML)
						jsonSettings.add(newJSON)
					}
				}

				return nil
			}); err != nil {
				return fmt.Errorf("error generating schema files: %w", err)
			}

			if len(yamlSettings) > 0 {
				yamlout, err := yamlSettings.marshal()
				if err != nil {
					return err
				}
				cmd.Println("Add the following to VS Code's settings.json file to enable YAML file validation:\n\n" + yamlout + "\n")
			}

			if len(jsonSettings) > 0 {
				jsonout, err := jsonSettings.marshal()
				if err != nil {
					return err
				}
				cmd.Println("Add the following to VS Code's settings.json file to enable JSON file validation:\n\n" + jsonout + "\n")
			}

			return nil
		},
	}

	return schemaCmd
}

func copyFile(srcFS fs.FS, dstDir, path string) error {
	src, err := srcFS.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file %q: %w", path, err)
	}

	destFile := filepath.Join(dstDir, path)

	dst, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("could not create file %q: %w", destFile, err)
	}

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("could not copy content to %q: %w", dst.Name(), err)
	}

	return nil
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
