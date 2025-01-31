package termdoc

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/codefmt"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/mdfmt"
)

// AutoMarkdownFormat produces the default terminal markdown formatter.
func AutoMarkdownFormat() *mdfmt.Formatter {
	columnsVal := TerminalWidth(120) // compute AOT
	codeFormatter := AutoCodeFormat()
	return &mdfmt.Formatter{
		// bold green with markdown header preserved
		Header: func(text string, loc mdfmt.Location) string {
			return green().Bold().Styled(
				fmt.Sprintf("%s %s",
					strings.Repeat("#", loc.Level),
					text,
				),
			)
		},
		Link: func(text, url string, loc mdfmt.Location) string {
			if loc.Header {
				// Do not change boldness of headers
				return fmt.Sprintf("%s%s",
					"["+text+"]",
					faint().Styled("("+url+")"))
			}
			return fmt.Sprintf("%s%s",
				bold().Styled("["+text+"]"),
				faint().Styled("("+url+")"))
		},
		Code: func(code string, loc mdfmt.Location) string {
			if loc.Header {
				return code
			}
			return cyan().Styled(code)
		},
		CodeBlock: func(code string, loc mdfmt.Location) string {
			switch loc.CodeBlockLang {
			case "bash", "sh", "python":
				return codeFormatter.Format(code, codefmt.LangInfo{
					LineCommentStart: "#",
				})
			case "go":
				return codeFormatter.Format(code, codefmt.LangInfo{
					LineCommentStart: "//",
				})
			default:
				return code
			}
		},
		Bold: func(text string, loc mdfmt.Location) string {
			if loc.Header {
				return text
			}
			return bold().Styled(text)
		},
		Italics: func(text string, loc mdfmt.Location) string {
			if loc.Header {
				return text
			}
			return italic().Styled(text)
		},
		Columns: func() int {
			return columnsVal
		},
		CodeBlockWrapMode: mdfmt.WrapToCurrentIndentation,
		Indent: func(loc mdfmt.Location) string {
			level := loc.Level
			if loc.Header {
				level-- // reduce by 1 for headers
			}
			switch level {
			case 0, 1:
				return ""
			default:
				return strings.Repeat("  ", level-1)
			}
		},
	}
}

// AutoCodeFormat produces the default terminal code formatter.
func AutoCodeFormat() *codefmt.Formatter {
	columnsVal := TerminalWidth(120) // compute AOT
	return &codefmt.Formatter{
		Comment: func(comment string, loc codefmt.Location) string {
			return faint().Styled(comment)
		},
		Columns: func() int {
			return columnsVal
		},
		WrapMode: codefmt.WrapToCurrentIndentation,
	}
}

//nolint:unused
var (
	style     = func(s ...string) termenv.Style { return termenv.DefaultOutput().String(s...) }
	bold      = func() termenv.Style { return style().Bold() }
	italic    = func() termenv.Style { return style().Italic() }
	underline = func() termenv.Style { return style().Underline() }
	faint     = func() termenv.Style { return style().Faint() }
	red       = func() termenv.Style { return style().Foreground(termenv.ANSIRed) }
	yellow    = func() termenv.Style { return style().Foreground(termenv.ANSIYellow) }
	green     = func() termenv.Style { return style().Foreground(termenv.ANSIGreen) }
	blue      = func() termenv.Style { return style().Foreground(termenv.ANSIBlue) }
	magenta   = func() termenv.Style { return style().Foreground(termenv.ANSIMagenta) }
	cyan      = func() termenv.Style { return style().Foreground(termenv.ANSICyan) }
)
