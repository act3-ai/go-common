// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/spf13/cobra"

	commands "git.act3-ace.com/ace/go-common/pkg/cmd"
	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/runner"
	vv "git.act3-ace.com/ace/go-common/pkg/version"
)

// manpages and schema definitions are embedded here for use in the gendocs and genschema commands
//
//go:embed manpages/* schemas/*
var embeds embed.FS

// getVersionInfo retreives the proper version information for this executable
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

func main() {
	info := getVersionInfo()

	// NOTE Often the main command is created elsewhere and imported
	root := &cobra.Command{
		Use: "sample",
	}

	// Create fs.FS rooted in the "manpages" dir
	manpages, err := fs.Sub(embeds, "manpages")
	if err != nil {
		log.Fatal(err)
	}

	// Create fs.FS rooted in the "schemas" dir
	schemas, err := fs.Sub(embeds, "schemas")
	if err != nil {
		log.Fatal(err)
	}

	schemaAssociations := []commands.SchemaAssociation{
		{
			Definition: "configuration-schema.json",
			FileMatch:  config.DefaultConfigValidatePath("ace", "sample", "config.yaml"),
		},
	}

	root.AddCommand(
		commands.NewVersionCmd(info),
		commands.NewGendocsCmd(manpages),
		commands.NewGenschemaCmd(schemas, schemaAssociations),
	)

	if err := runner.Run(root, "ACE_SAMPLE_VERBOSITY"); err != nil {
		// fmt.Fprintln(os.Stderr, "Error occurred", err)
		os.Exit(1)
	}
}
