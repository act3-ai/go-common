package termdoc

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/mdfmt"
)

// AutoColorFormat produces the default format.
func AutoColorFormat() *mdfmt.Formatter {
	columnsVal := TerminalWidth(120) // compute AOT
	return &mdfmt.Formatter{
		// bold green with markdown header preserved
		Header: func(text string, loc mdfmt.MDLocation) string {
			return green().Bold().Styled(
				fmt.Sprintf("%s %s",
					strings.Repeat("#", loc.Level),
					text,
				),
			)
		},
		Link: func(text, url string, loc mdfmt.MDLocation) string {
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
		Code: func(code string, loc mdfmt.MDLocation) string {
			if loc.Header {
				return code
			}
			return cyan().Styled(code)
		},
		Bold: func(text string, loc mdfmt.MDLocation) string {
			if loc.Header {
				return text
			}
			return bold().Styled(text)
		},
		Italics: func(text string, loc mdfmt.MDLocation) string {
			if loc.Header {
				return text
			}
			return italic().Styled(text)
		},
		Columns: func() int {
			return columnsVal
		},
		Indent: func(loc mdfmt.MDLocation) string {
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
