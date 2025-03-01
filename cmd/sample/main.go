// Package main is a sample CLI tool to demonstrate how these libraries are properly composed.
package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/cobrautil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/optionshelp"
	"gitlab.com/act3-ai/asce/go-common/pkg/otel"
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
//go:embed docs/*.md
var docs embed.FS

// getVersionInfo retreives the proper version information for this executable
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

func mainSetup() (context.Context, *cobra.Command, *otel.Config, error) {
	info := getVersionInfo()        // Load the version info from the build
	root := newSample(info.Version) // Create the root command

	ctx := context.Background()

	r, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName("sample"),
			semconv.ServiceVersion(info.Version),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("OTEL resource setup failed: %w", err)
	}

	otelCfg := &otel.Config{
		Resource: r,
		// Hardcoded exporters may be added here...
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
	return ctx, root, otelCfg, nil
}

func mainE(args []string) error {
	ctx, root, otelCfg, err := mainSetup()
	if err != nil {
		return err
	}
	root.SetArgs(args)

	// Run root command with OTel instrumentation enabled.
	return otel.Run(ctx, root, otelCfg, "ACE_SAMPLE_VERBOSITY")
}

func main() {
	if err := mainE(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

//go:embed docs/testfile.md
var testFile string

// Define command options.
type sampleAction struct {
	greeting string
	name     string
	count    int
	excited  bool
}

// NOTE Often the main command is created in another package and imported
func newSample(version string) *cobra.Command {
	action := &sampleAction{}

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
			if action.name == "" {
				action.name = "world"
			}
			suffix := ""
			if action.excited {
				suffix = "!"
			}
			for range action.count {
				cmd.Println(action.greeting + " " + action.name + suffix)
			}
		},
	}

	// Add flags to the command
	optionGroups := addFlags(root.Flags(), action)

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
				return termdoc.AutoCodeFormat().Format(s, codefmt.Bash)
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

	termMD := termdoc.AutoMarkdownFormat()

	root.AddCommand(
		optionshelp.Command(
			"sample-config", "Help for sample CLI configuration", optionGroups, termMD),
		// Add "Additional Help Topic" command that simply prints documentation.
		termdoc.AdditionalHelpTopic(
			"testfile", "Help command that displays the test file", testFile, termMD),
	)

	return root
}

func addFlags(f *pflag.FlagSet, action *sampleAction) []*options.Group {
	// Define each option
	name := &options.Option{
		Type:          options.String,
		Default:       "",
		JSON:          "name",
		Env:           "ACE_SAMPLE_NAME", // flagutil.ParseEnvOverrides uses this to set the value.
		Flag:          "name",
		FlagShorthand: "n",
		Short:         "Your name.",
		Long: heredoc.Doc(`
			Name of the sample CLI's user.`),
	}
	greeting := &options.Option{
		Type:          options.String,
		Default:       "Hello",
		JSON:          "",
		Env:           "ACE_SAMPLE_GREETING",
		Flag:          "greeting",
		FlagShorthand: "g",
		Short:         "Greeting for the user.",
		Long:          ``,
	}
	count := &options.Option{
		Type:          options.Integer,
		Default:       "1",
		JSON:          "",
		Env:           "ACE_SAMPLE_COUNT",
		Flag:          "count",
		FlagShorthand: "c",
		Short:         "Number of greetings to output.",
		Long:          ``,
	}
	excited := &options.Option{
		Type:          options.Boolean,
		Default:       "false",
		JSON:          "",
		Env:           "ACE_SAMPLE_EXCITED",
		Flag:          "excited",
		FlagShorthand: "e",
		Short:         "Greet with excitement.",
		Long:          ``,
	}

	// Create a group for the options
	group := &options.Group{
		Name:        "example",
		Description: "Example options",
		Options: []*options.Option{
			name,
			greeting,
			count,
			excited,
		},
	}

	// Add flags to the command.
	nameFlag := options.StringVar(f, &action.name, "", name)
	greetingFlag := options.StringVar(f, &action.greeting, "Hello", greeting)
	countFlag := options.IntVar(f, &action.count, 1, count)
	excitedFlag := options.BoolVar(f, &action.excited, false, excited)

	// Mark the flags as grouped to organize in help text.
	options.GroupFlags(group,
		nameFlag,
		greetingFlag,
		countFlag,
		excitedFlag,
	)

	return []*options.Group{group}
}
