// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/cobrautil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/codefmt"
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
	info := getVersionInfo()        // Load the version info from the build
	root := newSample(info.Version) // Create the root command

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
		os.Exit(1)
	}
}

//go:embed docs/testfile.md
var testFile string

// NOTE Often the main command is created in another package and imported
func newSample(version string) *cobra.Command {
	// Flag variable declaractions.
	var (
		greeting string
		name     string
		count    int
		excited  bool
	)

	root := &cobra.Command{
		Use: "sample",
		Example: heredoc.Doc(`
			# Run sample:
			sample
			
			# Run sample with name set by flag:
			sample --name "Foo"
			
			# Run sample with name set by environment variable:
			ACE_SAMPLE_NAME="Foo" sample`),
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse environment variables to set the value of flags,
			// if they have an environment variable defined.
			// Environment variables can be set with flagutil.SetEnvName.
			// The flag creation functions in pkg/options/flags.go set an
			// environment variable for the flag if Option.Env is set.
			return cobrautil.ParseEnvOverrides(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				name = "world"
			}
			suffix := ""
			if excited {
				suffix = "!"
			}
			for range count {
				cmd.Println(greeting + " " + name + suffix)
			}
		},
	}

	// Add flags to the command.
	nameFlag := options.StringVar(root.Flags(), &name, "",
		&options.Option{
			Type:          options.String,
			Default:       "",
			Path:          "name",
			Env:           "ACE_SAMPLE_NAME", // flagutil.ParseEnvOverrides uses this to set the value.
			Flag:          "name",
			FlagShorthand: "n",
			Short:         "Your name.",
			Long: heredoc.Doc(`
				Name of the sample CLI's user.`),
		})
	greetingFlag := options.StringVar(root.Flags(), &greeting, "Hello",
		&options.Option{
			Type:          options.String,
			Default:       "Hello",
			Path:          "",
			Env:           "ACE_SAMPLE_GREETING",
			Flag:          "greeting",
			FlagShorthand: "g",
			Short:         "Greeting for the user.",
			Long:          ``,
		})
	countFlag := options.IntVar(root.Flags(), &count, 1,
		&options.Option{
			Type:          options.Integer,
			Default:       "1",
			Path:          "",
			Env:           "ACE_SAMPLE_COUNT",
			Flag:          "count",
			FlagShorthand: "c",
			Short:         "Number of greetings to output.",
			Long:          ``,
		})
	excitedFlag := options.BoolVar(root.Flags(), &excited, false,
		&options.Option{
			Type:          options.Boolean,
			Default:       "false",
			Path:          "",
			Env:           "ACE_SAMPLE_EXCITED",
			Flag:          "excited",
			FlagShorthand: "e",
			Short:         "Greet with excitement.",
			Long:          ``,
		})

	// Create a group to organize flags in help text.
	options.GroupFlags(
		&options.Group{
			Name:        "example",
			Description: "Example options",
		},
		nameFlag,
		greetingFlag,
		countFlag,
		excitedFlag,
	)

	// Formatting options to style the help command text.
	//
	// These are just an example of some styling and content choices that can be made.
	formatOptions := cobrautil.UsageFormatOptions{
		Format: cobrautil.Formatter{
			// Format headers as uppercase.
			Header: strings.ToUpper,
			// Format command examples as Bash code snippets
			// with faint comments.
			Example: func(s string) string {
				return codeFormatter.Format(s, codefmt.Bash)
			},
		},
		// Options for the display of flag usages.
		FlagOptions: flagutil.UsageFormatOptions{
			// Override flag type name with type name configured by `options` package.
			FormatType: func(flag *pflag.Flag, typeName string) string {
				opt := options.FromFlag(flag)
				if opt.FlagType != "" {
					typeName = opt.FlagType
				}
				return typeName
			},
			// If there is a configured environment variable for the flag,
			// add environment variable name to the usage string
			FormatUsage: func(flag *pflag.Flag, usage string) string {
				opt := options.FromFlag(flag)
				if opt.Env != "" {
					usage += fmt.Sprintf(" (env: %s)", opt.Env)
				}
				return usage
			},
		},
		// Set local flags to be separated by grouping.
		LocalFlags: cobrautil.FlagGroupingOptions{
			GroupFlags: true,
		},
	}

	// Set custom usage function to format command
	// help text using our special formatting.
	cobrautil.WithCustomUsage(root, formatOptions)

	// Set custom formatting for the gendocs command
	// so generated docs match the format of the help text.
	embedutil.SetUsageFormat(formatOptions)

	root.AddCommand(
		// Add "Additional Help Topic" command that simply prints documentation.
		termdoc.AdditionalHelpTopic("testfile", "Help command that displays the test file", testFile),
	)

	return root
}

var codeFormatter = codefmt.Formatter{
	Comment: func(comment string, loc codefmt.Location) string {
		return termenv.String().Faint().Styled(comment)
	},
}
