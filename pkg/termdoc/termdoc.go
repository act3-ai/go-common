package termdoc

import (
	"fmt"
	"os"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/term"
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
		return "`" + s + "`"
	}
	return s
}

// CodeBlock renders a code block as Markdown if color output is disabled.
func CodeBlock(language, s string) string {
	if noColor() {
		// Return Markdown-formatted
		return "\n```" + language + "\n" + strings.TrimSuffix(s, "\n") + "\n```"
	}
	return s
}

// Footer renders a footer as Markdown if color output is disabled.
func Footer(s string) string {
	if noColor() {
		// Return Markdown-formatted
		lines := []string{}
		for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
			if strings.TrimSpace(line) == "" {
				lines = append(lines, ">") // no trailing space for lint reasons
			} else {
				lines = append(lines, "> "+line) // no trailing space for lint reasons
			}
		}
		return strings.Join(lines, "\n")
	}
	return s
}

// UList renders an unordered list as Markdown if color output is disabled.
func UList(defaultPrefix string, items ...string) string {
	if noColor() {
		// Return Markdown-formatted with starting newline
		result := "\n"
		for _, item := range items {
			result += "- " + item + "\n"
		}
		return strings.TrimSuffix(result, "\n")
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
		result := "\n"
		for i, item := range items {
			result += fmt.Sprintf("%d. %s\n", i+1, item)
		}
		return strings.TrimSuffix(result, "\n")
	}

	// Return formatted with no leading newline
	result := ""
	for i, item := range items {
		result += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	return strings.TrimSuffix(result, "\n")
}
