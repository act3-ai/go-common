/*
Package genschema generates JSON Schema definitions for Go types.

The JSON Schema generation uses [invopop/jsonschema], which is based on type reflection. The schema definitions are intended to be embedded in a Go CLI binary to be "generated" on the user's system.

# Example

Below is an example of how to generate the schema use the go:generate directive:

	// gen.go
	package gen

	//go:generate go run internal/gen/main.go cmd/example/schemas

And the file called by the go:generate directive in gen.go:

	// internal/gen/main.go

	//go:build ignore

	package main

	import (
		"fmt"
		"log"

		"github.com/act3-ai/go-common/pkg/genschema"
		"git.act3-ace.com/ace/example/pkg/apis/example.act3-ace.io/v1alpha1"
	)

	func main() {
		if len(os.Args) < 1 {
			log.Fatal("Must specify a target directory for schema generation.")
		}
	 	// Generate JSON Schema definitions
	 	if err := genschema.GenJSONSchema(
	 		"cmd/act3-pt/schemas",
	 		[]any{&v1alpha1.Configuration{}, &v1alpha1.Data{}},
	 		"example.act3-ace.io/v1alpha1",
	 		"git.act3-ace.com/ace/example",
	 	); err != nil {
	 		log.Fatal(fmt.Errorf("JSON Schema generation failed: %w", err))
	 	}
	}

And finally, embedding the JSON Schema definitions in a CLI and adding the "genschema" command:

	// cmd/example/main.go
	package main

	import (
		"embed"
		"io/fs"
		"log"
		"os"

		"github.com/spf13/cobra"

		commands "github.com/act3-ai/go-common/pkg/cmd"
	)

	//go:embed schemas/*
	var schemas embed.FS

	func main() {
		cmd := &cobra.Command{
			Use: "example",
		}

		schemaAssociations := []SchemaAssociation{
			{
				Definition: "schemas/configuration-schema.json",
				FileMatch:  []string{"ace-example-configuration.yaml"},
			},
			{
				Definition: "schemas/data-shema.json",
				FileMatch:  []string{"ace-example-data.json"},
			},
		}

		cmd.AddCommand(
			commands.NewGenschemaCmd(schemas, schemaAssociations),
		)

		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
	}

Now, running "go generate ./..." before running "go build ./cmd/example" results in a CLI with a "genschema" command that will generate accurate JSON Schema definitions for the provided schemas.
*/
package genschema
