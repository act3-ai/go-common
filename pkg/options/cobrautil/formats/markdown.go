package formats

import (
	"fmt"
	"strings"

	"github.com/act3-ai/go-common/pkg/options/cobrautil"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
	"github.com/spf13/pflag"
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
func Markdown() cobrautil.UsageFormatOptions {
	return cobrautil.UsageFormatOptions{
		Format:      markdownCobraFormatter(),
		FlagOptions: markdownFlagFormatter(),
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
}

func markdownCobraFormatter() cobrautil.Formatter {
	return cobrautil.Formatter{
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
	}
}

func markdownFlagLineFunc(flag *pflag.Flag) (line string, skip bool) {
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
}

func markdownFlagFormatter() flagutil.UsageFormatOptions {
	return flagutil.UsageFormatOptions{
		// Disable column wrapping
		Columns: flagutil.StaticColumns(0),
		// Custom line function creating markdown output.
		LineFunc: markdownFlagLineFunc,
	}
}
