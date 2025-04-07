package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/act3-ai/go-common/pkg/embedutil"
	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/options/cobrautil"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
	"github.com/act3-ai/go-common/pkg/options/optionshelp"
	"github.com/act3-ai/go-common/pkg/termdoc"
	"github.com/act3-ai/go-common/pkg/termdoc/codefmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
