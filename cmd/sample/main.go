// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"embed"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/cobrautil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
	vv "gitlab.com/act3-ai/asce/go-common/pkg/version"
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

	var name string

	// NOTE Often the main command is created elsewhere and imported
	root := &cobra.Command{
		Use: "sample",
		Example: heredoc.Doc(`
			# Run sample:
			sample
			
			# Run sample with name flag:
			sample --name "Foo"`),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Hello " + name)
		},
	}

	// Disable flag sorting
	cobrautil.WalkCommands(root, func(cmd *cobra.Command) {
		cmd.Flags().SortFlags = false
	})

	// Formatting options to style the help command text.
	formatOptions := cobrautil.UsageFormatOptions{
		Format: cobrautil.Formatter{
			// Format headers as uppercase.
			Header: strings.ToUpper,
		},
		// Options for the display of flag usages.
		FlagOptions: flagutil.UsageFormatOptions{
			FormatType: func(flag *pflag.Flag, typeName string) string {
				return strings.ToLower(typeName)
			},
		},
		// Set local flags to be separated by grouping.
		LocalFlags: cobrautil.FlagGroupingOptions{
			GroupFlags: true,
		},
	}

	// Set custom usage function.
	cobrautil.WithCustomUsage(root, formatOptions)

	// Set custom formatting for the gendocs command.
	embedutil.SetUsageFormat(formatOptions)

	nameFlag := options.StringVar(root.Flags(), &name, "",
		&options.Option{
			Type:          options.String,
			Default:       "",
			Path:          "name",
			Env:           "ACE_SAMPLE_NAME",
			Flag:          "name",
			FlagShorthand: "n",
			Short:         "Your name.",
			Long: heredoc.Doc(`
				Name of the sample CLI's user.`),
		})

	options.GroupFlags(
		&options.Group{
			Name:        "example",
			Description: "Example options",
		},
		nameFlag,
	)

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		name = flagutil.ValueOr(nameFlag, name, "Sample User")
	}

	schemaAssociations := []commands.SchemaAssociation{
		{
			Definition: "configuration-schema.json",
			FileMatch:  config.DefaultConfigValidatePath("ace", "sample", "config.yaml"),
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

	if err := runner.Run(root, "ACE_SAMPLE_VERBOSITY"); err != nil {
		// fmt.Fprintln(os.Stderr, "Error occurred", err)
		os.Exit(1)
	}
}
