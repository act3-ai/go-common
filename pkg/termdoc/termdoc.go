package termdoc

import (
	"fmt"
	"os"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/term"

	"github.com/act3-ai/go-common/pkg/md"
)

// TerminalWidth returns the width of the terminal, using fallback if it can't determine width.
func TerminalWidth(fallback int) int {
	w := termenv.DefaultOutput().Writer()
	if w == nil {
		return fallback
	}
	f, ok := w.(*os.File)
	if !ok {
		return fallback
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return fallback
	}
	return width
}

// noColor reports whether color output is enabled.
func noColor() bool {
	return termenv.DefaultOutput().Profile == termenv.Ascii ||
		termenv.EnvNoColor()
}

// Header renders a header as a Markdown h3 if color output is disabled.
func Header(s string) string {
	if noColor() {
		// Return Markdown-formatted
		return "### " + strings.TrimSuffix(s, ":") + "\n"
	}
	return s
}

// Code renders an inline Code block as Markdown if color output is disabled.
func Code(s string) string {
	if noColor() {
		// Return Markdown-formatted
		return md.Code(s)
	}
	return s
}

// CodeBlock renders a code block as Markdown if color output is disabled.
func CodeBlock(language, s string) string {
	if noColor() {
		// Return Markdown-formatted
		return "\n" + md.CodeBlock(language, strings.TrimSuffix(s, "\n"))
	}
	return s
}

// Footer renders a footer as Markdown if color output is disabled.
func Footer(s string) string {
	if noColor() {
		// Return Markdown-formatted
		return md.BlockQuote(strings.TrimSpace(s))
	}
	return s
}

// UList renders an unordered list as Markdown if color output is disabled.
func UList(defaultPrefix string, items ...string) string {
	if noColor() {
		// Return Markdown-formatted with starting newline
		return "\n" + md.UList(items...)
	}

	// Return formatted with default prefix
	result := ""
	for _, item := range items {
		result += defaultPrefix + item + "\n"
	}
	return strings.TrimSuffix(result, "\n")
}

// OList renders an ordered list as Markdown if color output is disabled.
func OList(items ...string) string {
	if noColor() {
		// Return Markdown-formatted with starting newline
		return "\n" + md.OList(items...)
	}

	// Return formatted with no leading newline
	result := ""
	for i, item := range items {
		result += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	return strings.TrimSuffix(result, "\n")
}
