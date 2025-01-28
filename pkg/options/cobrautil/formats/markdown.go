package formats

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/cobrautil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
)

//nolint:unused
var (
	mdBold      = func(s string) string { return "__" + s + "__" }
	mdUnderline = func(s string) string { return "<u>" + s + "</u>" }
	mdItalics   = func(s string) string { return "_" + s + "_" }
	mdCode      = func(s string) string { return "`" + s + "`" }
	mdCodeBlock = func(lang, s string) string { return "```" + lang + "\n" + s + "\n```" }
)

// Markdown is a format producing valid markdown.
var Markdown = cobrautil.UsageFormatOptions{
	Format: cobrautil.Formatter{
		// Formats headers as markdown h2
		Header: func(s string) string {
			return "## " + strings.TrimSuffix(s, ":") + "\n"
		},
		// Format commands bold
		Command: func(s string) string {
			return mdBold(s)
		},
		// Formats argument placeholders bold
		Args: func(s string) string {
			return mdBold(s)
		},
		// Formats command snippets as inline code snippets
		CommandAndArgs: func(s string) string {
			return mdCode(s)
		},
		// Formats examples as bash code blocks
		Example: func(s string) string {
			return mdCodeBlock("bash", s)
		},
	},
	FlagOptions: flagutil.UsageFormatOptions{
		// Disable column wrapping
		Columns: flagutil.StaticColumns(0),
		// Custom line function creating markdown output.
		LineFunc: func(flag *pflag.Flag) (line string, skip bool) {
			if flag.Hidden || flag.Deprecated != "" {
				return "", true
			}

			var (
				flagName  string
				flagType  string
				flagUsage string
			)

			if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
				flagName = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
			} else {
				flagName = "--" + flag.Name
			}

			flagType, flagUsage = pflag.UnquoteUsage(flag)
			flagUsage = strings.ReplaceAll(flagUsage, "\n", "\n  ")

			envName := flagutil.GetEnvName(flag)
			if envName != "" {
				flagUsage += fmt.Sprintf(" (env: %s)", mdCode(envName))
			}

			if !flagutil.DefaultIsZeroValue(flag) {
				defValue := flag.DefValue
				if flag.Value.Type() == "string" {
					defValue = fmt.Sprintf("%q", defValue)
				}
				flagUsage += fmt.Sprintf(" (default %s)", mdBold(defValue))
			}

			flagName = mdCode(flagName)
			if flagType != "" {
				flagType = " " + mdItalics(flagType)
			}

			// Create an unordered list entry
			line = fmt.Sprintf("- %s%s: %s", flagName, flagType, flagUsage)

			return line, false
		},
	},
	LocalFlags: cobrautil.FlagGroupingOptions{
		// Separate local flags into groups,
		// if defined by [options.GroupFlags].
		GroupFlags: true,
	},
	InheritedFlags: cobrautil.FlagGroupingOptions{
		// Separate inherited flags into groups,
		// if defined by [options.GroupFlags].
		GroupFlags: true,
	},
}

var mdFlagOptions = flagutil.UsageFormatOptions{
	// Disable column wrapping
	Columns: flagutil.StaticColumns(0),
	// Format flag name as inline code snippet
	FormatFlagName: func(flag *pflag.Flag, name string) string {
		return mdCode(name)
	},
	// Formats flag type in italics and overrides plural slice types with ellipses
	FormatType: func(flag *pflag.Flag, typeName string) string {
		opt := options.FromFlag(flag)
		if opt.FlagType != "" {
			typeName = opt.FlagType
		}
		// Override plural slice types with ellipses
		switch typeName {
		case "strings":
			typeName = "string..."
		case "ints":
			typeName = "int..."
		case "uints":
			typeName = "uint..."
		case "bools":
			typeName = "bool..."
		}
		return mdItalics(typeName)
	},
	// Formats flag values bold
	FormatValue: func(flag *pflag.Flag, value string) string {
		return mdBold(value)
	},
	// Adds environment variable name to the usage string,
	// (if set by [flagutil.SetEnvName])
	// Formats environment variable name as inline code snippet (to match flag name)
	FormatUsage: func(flag *pflag.Flag, usage string) string {
		envName := flagutil.GetEnvName(flag)
		if envName != "" {
			usage += fmt.Sprintf(" (env: %s)", mdCode(envName))
		}
		return usage
	},
}
