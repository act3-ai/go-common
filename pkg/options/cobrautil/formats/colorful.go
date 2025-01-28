// Package formats defines some example formats for use with the cobrautil and flagutil packages.
package formats

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
	"github.com/spf13/pflag"
	"gitlab.com/act3-ai/asce/go-common/pkg/options"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/cobrautil"
	"gitlab.com/act3-ai/asce/go-common/pkg/options/flagutil"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc"
)

//nolint:unused
var (
	ansiStyle     = func(s ...string) termenv.Style { return termenv.DefaultOutput().String(s...) }
	ansiBold      = func() termenv.Style { return ansiStyle().Bold() }
	ansiUnderline = func() termenv.Style { return ansiStyle().Underline() }
	ansiFaint     = func() termenv.Style { return ansiStyle().Faint() }
	ansiRed       = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIRed) }
	ansiYellow    = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIYellow) }
	ansiGreen     = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIGreen) }
	ansiBlue      = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIBlue) }
	ansiMagenta   = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIMagenta) }
	ansiCyan      = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSICyan) }
)

// Colorful is a colorful formatting option.
var Colorful = cobrautil.UsageFormatOptions{
	Format: cobrautil.Formatter{
		// Formats headers bold green
		Header: func(s string) string {
			return ansiGreen().Bold().Styled(s)
		},
		// Formats commands bold cyan
		Command: func(s string) string {
			return ansiCyan().Bold().Styled(s)
		},
		// Formats argument placeholders cyan
		Args: func(s string) string {
			return ansiCyan().Styled(s)
		},
		// Formats examples like bash code snippets
		// Comments (lines beginning with "#") are faint
		Example: func(s string) string {
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				switch {
				// Format bash comments with faint text
				case strings.HasPrefix(strings.TrimSpace(line), "#"):
					lines[i] = ansiStyle().Faint().Styled(line)
				// Leave commands as-is
				default:
					lines[i] = line
				}
			}
			return strings.Join(lines, "\n")
		},
	},
	FlagOptions: flagutil.UsageFormatOptions{
		// Columns are set by dynamically querying the output writer for terminal width.
		// If output is not a TTY, the value 80 will be used.
		Columns: flagutil.DynamicColumns(func() int {
			return termdoc.TerminalWidth(80)
		}),
		// Formats flag name bold cyan
		FormatFlagName: func(flag *pflag.Flag, name string) string {
			return ansiCyan().Bold().Styled(name)
		},
		// Formats flag type cyan and override plural slice types with ellipses
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
			return ansiCyan().Styled(typeName)
		},
		// Formats flag values bold
		FormatValue: func(flag *pflag.Flag, value string) string {
			return ansiBold().Styled(value)
		},
		// Adds environment variable name to the usage string,
		// (if set by [flagutil.SetEnvName])
		// Formats environment variable name bold cyan (to match flag name)
		FormatUsage: func(flag *pflag.Flag, usage string) string {
			envName := flagutil.GetEnvName(flag)
			if envName != "" {
				usage += fmt.Sprintf(" (env: %s)", ansiCyan().Bold().Styled(envName))
			}
			return usage
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
