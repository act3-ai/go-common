package mdfmt

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Markdown component regexes
var (
	mdBoldUnderlineRegex   = regexp.MustCompile(wordish(`__([^_]+)__`))      // __bold__
	mdBoldAsteriskRegex    = regexp.MustCompile(wordish(`\*\*([^\*]+)\*\*`)) // **bold**
	mdItalicUnderlineRegex = regexp.MustCompile(wordish(`_([^_]+)_`))        // _italic_
	mdItalicAsteriskRegex  = regexp.MustCompile(wordish(`\*([^\*]+)\*`))     // *italic*
	mdCodeRegex            = regexp.MustCompile("`[^`]+`")                   // `code`
	mdLinkRegex            = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)  // [text](url)

	startWord = `^(?:.*\s)?`
	endWord   = `(?:\s.*)?$`
)

const (
	codeBlockStart = "```"
	commentStart   = "<!--"
	commentEnd     = "-->"
)

func wordish(re string) string {
	return startWord + re + endWord
}

// Format formats markdown text according the Formatter's rules.
func (format *Formatter) Format(markdownText string) string {
	cols := 0
	if format.Columns != nil {
		cols = format.Columns()
	}

	lines := strings.Split(markdownText, "\n")
	formatted := make([]string, 0, len(lines))
	var loc MDLocation
	codeBlockIndent := ""
	codeBlockStop := ""
	for _, line := range lines {
		lineTrimSpace := strings.TrimSpace(line)
		switch {
		// In open comment, only check if exiting
		case loc.Comment:
			_, afterEnd, foundEnd := strings.Cut(line, commentEnd)
			if !foundEnd {
				continue // skip comment lines
			}
			// Exit comment
			loc.Comment = false
			line = afterEnd
			lineTrimSpace = strings.TrimSpace(afterEnd)
			if lineTrimSpace == "" {
				// skip if empty
				continue
			}
			// Format content after comment
			fallthrough
		// In code block, only check if exiting
		case loc.CodeBlock:
			// Exit code block
			if strings.HasPrefix(lineTrimSpace, codeBlockStop) {
				loc.CodeBlock = false
				loc.CodeBlockLevel = 0
				codeBlockStop = ""
			}
		// Start code block
		case strings.HasPrefix(lineTrimSpace, codeBlockStart):
			loc.CodeBlock = true
			loc.CodeBlockLevel = codeBlockLevel(lineTrimSpace)
			codeBlockStop = strings.Repeat("`", loc.CodeBlockLevel)
			codeBlockIndent = extraIndent(line)
		// Comment line
		case strings.Contains(line, commentStart):
			beforeStart, _, _ := strings.Cut(line, commentStart)
			_, afterEnd, foundEnd := strings.Cut(line, commentEnd)
			switch {
			// Start of multiline comment
			// Markdown comments are only multiline if the
			// comment starts at the beginning of the line
			case beforeStart == "" && !foundEnd:
				// start skipping comment lines
				loc.Comment = true
				continue
			// End of comment not found, but comment is not multiline
			// --or--
			// End of comment was found, which means the comment is not multiline
			default:
				line = beforeStart + afterEnd
				lineTrimSpace = strings.TrimSpace(line)
				if lineTrimSpace == "" {
					continue // skip empty comment surroundings
				}
			}
			fallthrough
		// Format non-code block line
		default:
			line = format.formatRegularLine(line, loc)
		}

		// Add section-defined indent:
		if format.Indent != nil {
			line = format.Indent(loc) + line
		}

		// Perform word wrapping:
		if cols > 0 {
			var indent string
			switch {
			// Preserve leading whitespace from where the codeblock was started
			case loc.CodeBlock:
				indent = codeBlockIndent
			// Preserve leading whitespace in the line
			default:
				indent = extraIndent(line)
			}
			// // Preserve leading whitespace from the line
			// indent := extraIndent(line)
			// Wrap lines
			line = ansi.Wordwrap(line, 100, " ")
			// Add indent to wrapped lines
			line = strings.ReplaceAll(line, "\n", "\n"+indent)
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

// Performs all non-word-wrap formatting for all markdown lines except those inside multiline code blocks
func (format *Formatter) formatRegularLine(line string, loc MDLocation) string {
	// Set header level (if header)
	if h := headerLevel(line); h > 0 {
		loc.Header = true
		loc.Level = h
	} else {
		loc.Header = false
	}

	// Markdown link formatter:
	if format.Link != nil {
		// Replace links first, the regex gets messed up by ANSI sequences
		line = mdLinkRegex.ReplaceAllStringFunc(line, func(s string) string {
			match := mdLinkRegex.FindStringSubmatch(s)
			return format.Link(match[1], match[2], loc)
		})
	}
	if format.Header != nil && loc.Header {
		line = format.Header(strings.TrimSpace(strings.TrimLeft(line, "#")), loc)
	}

	// Markdown inline code formatter:
	if format.Code != nil {
		line = mdCodeRegex.ReplaceAllStringFunc(line, func(s string) string {
			return format.Code(s[1:len(s)-1], loc)
		})
	}

	// Markdown bold formatter:
	if format.Bold != nil {
		// Replace both underline and asterisk notation
		line = mdBoldUnderlineRegex.ReplaceAllStringFunc(line,
			func(s string) string {
				return format.Bold(s[2:len(s)-2], loc)
			})
		line = mdBoldAsteriskRegex.ReplaceAllStringFunc(line,
			func(s string) string {
				return format.Bold(s[2:len(s)-2], loc)
			})
	}

	// Markdown italic formatter:
	if format.Italics != nil {
		// Replace both underline and asterisk notation
		line = mdItalicUnderlineRegex.ReplaceAllStringFunc(line,
			func(s string) string {
				return format.Italics(s[1:len(s)-1], loc)
			})
		line = mdItalicAsteriskRegex.ReplaceAllStringFunc(line,
			func(s string) string {
				return format.Italics(s[1:len(s)-1], loc)
			})
	}

	return line
}

func headerLevel(s string) int {
	if strings.HasPrefix(s, "#") {
		return 1 + headerLevel(strings.TrimPrefix(s, "#"))
	}
	return 0
}

func extraIndent(s string) string {
	if strings.HasPrefix(s, " ") {
		return " " + extraIndent(strings.TrimPrefix(s, " "))
	} else if strings.HasPrefix(s, "\t") {
		return "\t" + extraIndent(strings.TrimPrefix(s, "\t"))
	}
	return ""
}

func codeBlockLevel(s string) int {
	if strings.HasPrefix(s, "`") {
		return 1 + codeBlockLevel(strings.TrimPrefix(s, "`"))
	}
	return 0
}
