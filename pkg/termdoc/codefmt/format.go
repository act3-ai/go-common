package codefmt

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Format formats markdown text according the Formatter's rules.
func (format *Formatter) Format(codeText string, lang LangInfo) string {
	cols := 0
	if format.Columns != nil {
		cols = format.Columns()
	}

	lines := strings.Split(codeText, "\n")
	formatted := make([]string, 0, len(lines))
	var loc Location
	for _, line := range lines {
		lineComment := strings.Index(line, lang.LineCommentStart)
		// lcBefore, lcAfter, lcFound := strings.Cut(line, lang.LineCommentStart)
		switch {
		// Format line comment
		// case lcFound:
		case lineComment != -1:
			// Format code before comment
			lcBefore := line[:lineComment]
			if format.Code != nil {
				lcBefore = format.Code(lcBefore, Location{LineComment: false})
			}
			// Format comment
			lcAfter := line[lineComment:]
			if format.Comment != nil {
				lcAfter = format.Comment(lcAfter, Location{LineComment: true})
			}
			// Reassemble the line
			line = lcBefore + lcAfter
		// Format code line
		default:
			if format.Code != nil {
				line = format.Code(line, Location{LineComment: false})
			}
		}

		// Add formatter-defined indent:
		if format.Indent != nil {
			line = format.Indent(loc) + line
		}

		// Perform word wrapping:
		if cols > 0 {
			// Preserve leading whitespace from the line
			// Must be determined from the line itself
			indent := extraIndent(line)
			// Wrap lines
			line = ansi.Wordwrap(line, cols, " ")
			// Add indent to wrapped lines
			line = strings.ReplaceAll(line, "\n", "\n"+indent)
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

func extraIndent(s string) string {
	if strings.HasPrefix(s, " ") {
		return " " + extraIndent(strings.TrimPrefix(s, " "))
	} else if strings.HasPrefix(s, "\t") {
		return "\t" + extraIndent(strings.TrimPrefix(s, "\t"))
	}
	return ""
}
