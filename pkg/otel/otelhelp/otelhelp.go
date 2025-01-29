// Package otelhelp defines CLI help commands with OTel configuration docs.
package otelhelp

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

// Fetch the otel docs
// //go:generate ./fetch-otel-docs.sh otel-general.md otel-otlp-exporter.md

// Embed the otel docs
var (
	//go:embed otel-general.md
	otelGeneral string

	//go:embed otel-otlp-exporter.md
	otelExporter string
)

// GeneralHelpCmd creates a help command for general OpenTelemetry configuration.
func GeneralHelpCmd() *cobra.Command {
	return helpTextCommand(
		"otel-options",
		"Help for general OpenTelemetry configuration.",
		otelGeneral,
	)
}

// ExporterHelpCmd creates a help command for OpenTelemetry Protocol Exporter (OTLP) configuration.
func ExporterHelpCmd() *cobra.Command {
	return helpTextCommand(
		"otel-exporter-options",
		"Help for OpenTelemetry Protocol Exporter (OTLP) configuration.",
		otelExporter,
	)
}

func helpTextCommand(name, short, long string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(0),
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Println(fancyFormat(cmd.Long))
	})
	return cmd
}

// Markdown component regexes
var (
	mdBoldRegex = regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	// mdItalicsRegex = regexp.MustCompile(`_([^_]+)_`)
	mdCodeRegex = regexp.MustCompile("`[^`]+`")
	mdLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
)

func fancyFormat(text string) string {
	if termenv.DefaultOutput().ColorProfile() == termenv.Ascii {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		header := strings.HasPrefix(line, "#")
		// Replace links first, the regex gets messed up by ANSI sequences
		line = mdLinkRegex.ReplaceAllStringFunc(line, func(s string) string {
			match := mdLinkRegex.FindStringSubmatch(s)
			return fmt.Sprintf("[%s]%s",
				bold().Styled(match[1]),
				faint().Styled("("+match[2]+")"))
		})
		if header {
			line = green().Bold().Styled(line)
		}
		// line = mdHeaderRegex.ReplaceAllStringFunc(line, green().Bold().Styled)
		line = mdCodeRegex.ReplaceAllStringFunc(line, func(s string) string {
			if header {
				return s[1 : len(s)-1]
			}
			return cyan().Styled(s[1 : len(s)-1])
		})

		line = mdBoldRegex.ReplaceAllStringFunc(line, func(s string) string {
			return bold().Styled(s[2 : len(s)-2])
		})
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

//nolint:unused
var (
	style     = func(s ...string) termenv.Style { return termenv.DefaultOutput().String(s...) }
	bold      = func() termenv.Style { return style().Bold() }
	underline = func() termenv.Style { return style().Underline() }
	faint     = func() termenv.Style { return style().Faint() }
	red       = func() termenv.Style { return style().Foreground(termenv.ANSIRed) }
	yellow    = func() termenv.Style { return style().Foreground(termenv.ANSIYellow) }
	green     = func() termenv.Style { return style().Foreground(termenv.ANSIGreen) }
	blue      = func() termenv.Style { return style().Foreground(termenv.ANSIBlue) }
	magenta   = func() termenv.Style { return style().Foreground(termenv.ANSIMagenta) }
	cyan      = func() termenv.Style { return style().Foreground(termenv.ANSICyan) }
)
