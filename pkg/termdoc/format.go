package termdoc

import (
	"fmt"
	"strings"

	"github.com/act3-ai/go-common/pkg/termdoc/codefmt"
	"github.com/act3-ai/go-common/pkg/termdoc/mdfmt"
	"github.com/muesli/termenv"
)

// AutoMarkdownFormat produces the default terminal markdown formatter.
func AutoMarkdownFormat() *mdfmt.Formatter {
	columnsVal := TerminalWidth(120) // compute AOT
	codeFormatter := AutoCodeFormat()
	return &mdfmt.Formatter{
		// bold green with markdown header preserved
		Header: func(text string, loc mdfmt.Location) string {
			return ansiGreen().Bold().Styled(
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
					ansiFaint().Styled("("+url+")"))
			}
			return fmt.Sprintf("%s%s",
				ansiBold().Styled("["+text+"]"),
				ansiFaint().Styled("("+url+")"))
		},
		Code: func(code string, loc mdfmt.Location) string {
			if loc.Header {
				return code
			}
			return ansiCyan().Styled(code)
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
			return ansiBold().Styled(text)
		},
		Italics: func(text string, loc mdfmt.Location) string {
			if loc.Header {
				return text
			}
			return ansiItalic().Styled(text)
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
			return ansiFaint().Styled(comment)
		},
		Columns: func() int {
			return columnsVal
		},
		WrapMode: codefmt.WrapToCurrentIndentation,
	}
}

//nolint:unused
var (
	ansiStyle     = func(s ...string) termenv.Style { return termenv.DefaultOutput().String(s...) }
	ansiBold      = func() termenv.Style { return ansiStyle().Bold() }
	ansiItalic    = func() termenv.Style { return ansiStyle().Italic() }
	ansiUnderline = func() termenv.Style { return ansiStyle().Underline() }
	ansiFaint     = func() termenv.Style { return ansiStyle().Faint() }
	ansiRed       = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIRed) }
	ansiYellow    = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIYellow) }
	ansiGreen     = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIGreen) }
	ansiBlue      = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIBlue) }
	ansiMagenta   = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSIMagenta) }
	ansiCyan      = func() termenv.Style { return ansiStyle().Foreground(termenv.ANSICyan) }
)
