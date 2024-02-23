// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"embed"
	"os"

	"github.com/spf13/cobra"

	commands "git.act3-ace.com/ace/go-common/pkg/cmd"
	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/embedutil"
	"git.act3-ace.com/ace/go-common/pkg/runner"
	vv "git.act3-ace.com/ace/go-common/pkg/version"
)

// manpages and schema definitions are embedded here for use in the gendocs and genschema commands
//
//go:embed schemas/*
var schemas embed.FS

// an example quick start guide is embedded here
// for use in the "gendocs" and "info" commands
//
//go:embed docs/*
var docs embed.FS

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

	schemaAssociations := []commands.SchemaAssociation{
		{
			Definition: "configuration-schema.json",
			FileMatch:  config.DefaultConfigValidatePath("cli", "sample", "config.yaml"),
		},
	}

	docs := &embedutil.Documentation{
		Title:   "Sample command showing the use of go-common's utilities for CLI development",
		Command: root,
		Categories: []*embedutil.Category{
			embedutil.NewCategory(
				"docs", "General Documentation", root.Name(), 7,
				embedutil.LoadMarkdown("quick-start-guide", "Example Quick Start Guide", "docs/quick-start-guide.md", docs),
			),
		},
	}

	commands.AddGroupedCommands(root,
		&cobra.Group{
			ID:    "utils",
			Title: "Utility commands",
		},
		commands.NewVersionCmd(info),
		commands.NewInfoCmd(docs),
		commands.NewGendocsCmd(docs),
		commands.NewGenschemaCmd(schemas, schemaAssociations),
	)

	if err := runner.Run(root, "SAMPLE_VERBOSITY"); err != nil {
		os.Exit(1)
	}
}
