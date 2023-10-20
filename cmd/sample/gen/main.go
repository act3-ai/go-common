//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"git.act3-ace.com/ace/go-common/pkg/genschema"
)

// Configuration for the sample CLI
type Configuration struct {
	// Your name
	Name string `json:"name"`
}

func main() {
	if len(os.Args) < 1 {
		log.Fatal("Must specify a target directory for schema generation.")
	}

	// Generate JSON Schema definitions
	if err := genschema.GenerateTypeSchemas(
		os.Args[1],
		[]any{&Configuration{}},
		"sample.act3-ace.io/v1alpha1",
		"git.act3-ace.com/ace/go-common",
	); err != nil {
		log.Fatal(fmt.Errorf("JSON Schema generation failed: %w", err))
	}
}
