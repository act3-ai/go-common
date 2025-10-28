package formats

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/md"
	"github.com/act3-ai/go-common/pkg/options/cobrautil"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
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
			return md.Bold(s)
		},
		// Formats argument placeholders bold
		Args: func(s string) string {
			return md.Bold(s)
		},
		// Formats command snippets as inline code snippets
		CommandAndArgs: func(s string) string {
			return md.Code(s)
		},
		// Formats examples as bash code blocks
		Example: func(s string) string {
			return md.CodeBlock("bash", s)
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
		flagUsage += fmt.Sprintf(" (env: %s)", md.Code(envName))
	}

	if !flagutil.DefaultIsZeroValue(flag) {
		defValue := flag.DefValue
		if flag.Value.Type() == "string" {
			defValue = fmt.Sprintf("%q", defValue)
		}
		flagUsage += fmt.Sprintf(" (default %s)", md.Bold(defValue))
	}

	flagName = md.Code(flagName)
	if flagType != "" {
		flagType = " " + md.Italics(flagType)
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
