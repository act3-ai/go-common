// Package formats defines some example formats for use with the cobrautil and flagutil packages.
package formats

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
	"github.com/spf13/pflag"

	"github.com/act3-ai/go-common/pkg/options"
	"github.com/act3-ai/go-common/pkg/options/cobrautil"
	"github.com/act3-ai/go-common/pkg/options/flagutil"
	"github.com/act3-ai/go-common/pkg/termdoc"
	"github.com/act3-ai/go-common/pkg/termdoc/codefmt"
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

// faintCommentsCodeFormatter that formats comments faint
func faintCommentsCodeFormatter() *codefmt.Formatter {
	return &codefmt.Formatter{
		Comment: func(comment string, loc codefmt.Location) string {
			return ansiStyle().Faint().Styled(comment)
		},
	}
}

// Colorful is a colorful formatting option.
func Colorful() cobrautil.UsageFormatOptions {
	return cobrautil.UsageFormatOptions{
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
				return faintCommentsCodeFormatter().Format(s, codefmt.Bash)
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
				// Override slice type names
				switch typeName {
				case "strings", "stringSlice":
					typeName = "string..."
				case "ints", "intSlice":
					typeName = "int..."
				case "uints", "uintSlice":
					typeName = "uint..."
				case "bools", "boolSlice":
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
					envUsage := fmt.Sprintf("(env: %s)", ansiCyan().Bold().Styled(envName))
					if strings.Contains(usage, "\n") {
						usage += "\n" + envUsage
					} else {
						usage += " " + envUsage
					}
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
}
