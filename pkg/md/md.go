package md

import (
	"fmt"
	"strconv"
	"strings"
)

// Header creates a header.
func Header(n int, header string) string {
	return strings.Repeat("#", n) + " " + header
}

// Bold creates bold text.
func Bold(text string) string {
	return "__" + text + "__"
}

// Italics creates italicized text.
func Italics(text string) string {
	return "_" + text + "_"
}

// Underline creates underlined text.
func Underline(text string) string {
	return "<u>" + text + "</u>"
}

// Code creates an inline code block.
func Code(code string) string {
	if strings.Contains(code, "\n") {
		panic("cannot format multiline string as inline code block")
	}
	return "`" + code + "`"
}

// CodeBlock creates a multiline code block.
func CodeBlock(lang, code string) string {
	return "```" + lang + "\n" + code + "\n```"
}

// UList renders the items as an unordered list.
func UList(items ...string) string {
	return "- " + strings.Join(items, "\n- ") + "\n"
}

// OList renders the items as an ordered list.
func OList(items ...string) string {
	out := ""
	for i, item := range items {
		out += strconv.Itoa(i+1) + ". " + item + "\n"
	}
	return out
}

// BlockQuote renders the text in a block quote.
func BlockQuote(text string) string {
	lines := []string{}
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) == "" {
			lines = append(lines, ">") // no trailing space for lint reasons
		} else {
			lines = append(lines, "> "+line) // no trailing space for lint reasons
		}
	}
	return strings.Join(lines, "\n")
}

// Details creates an HTML <details> element.
//
//nolint:revive
func Details(summary, body string, open bool) string {
	var openStr string
	if open {
		openStr = ` open="true"`
	}
	return fmt.Sprintf(`<details%s>
<summary>%s</summary>

%s</details>`,
		openStr, summary, body)
}

// Link creates a link.
func Link(text, target string) string {
	return fmt.Sprintf("[%s](%s)", text, target)
}

// HeaderLinkTarget encodes the text in link target form.
func HeaderLinkTarget(header string) string {
	// Begin with target character "#"
	return "#" + toMarkdownLinkFragment(header)
}

// toMarkdownLinkFragment formats the string as a markdown link fragment.
func toMarkdownLinkFragment(s string) string {
	// Lowercase
	return strings.ToLower(
		// Replace forbidden characters
		mdLinkTargetReplacer.Replace(
			// Trim forbidden leading/trailing characters
			strings.Trim(s, mdlinkCutset)))
}

// mdlinkCutset is used to trim characters from the beginning and end of strings.
var mdlinkCutset = "-"

const zeroString = ""

// mdLinkTargetReplacer replaces characters to produce the equivalent markdown link handle
var mdLinkTargetReplacer = strings.NewReplacer(
	" ", "-",
	".", zeroString,
	"/", zeroString,
	"*", zeroString,
	"`", zeroString,
	"'", zeroString,
	`"`, zeroString,
)
