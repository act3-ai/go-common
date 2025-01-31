package codefmt

// Location describes the current location of text in a document.
type Location struct {
	LineComment bool // In a line comment
	// MultilineComment bool   // In a multiline comment
}

// LangInfo defines basic language information needed for parsing.
type LangInfo struct {
	LineCommentStart string // Starts line comments
	// MultilineCommentStart string // Starts multiline comments
	// MultilineCommentEnd   string // Ends multiline comments
}

// Defined LangInfo for reuse.
var (
	Bash = LangInfo{
		LineCommentStart: "#",
	}

	Go = LangInfo{
		LineCommentStart: "//",
		// MultilineCommentStart: "/*",
		// MultilineCommentEnd:   "*/",
	}
)

// Formatter formats Markdown for terminal output.
type Formatter struct {
	Comment func(comment string, loc Location) string // reformats inline code blocks
	Code    func(code string, loc Location) string    // reformats inline code blocks
	Indent  func(loc Location) string                 // produces indent for a line's location

	// produce column width for wrapping
	// (nil function or 0 return value disables wrapping)
	Columns func() int

	WrapMode WrapMode
}

// StaticColumns is a static columns setting.
func StaticColumns(cols int) func() int {
	return func() int {
		return cols
	}
}

// WrapMode signifies a code block wrapping style.
type WrapMode uint8

// Defined code block wrapping styles.
const (
	Default                   WrapMode = iota
	WrapToCurrentIndentation           // Indent wrapped lines to the current line's indentation
	WrapToStartingIndentation          // Indent wrapped lines to the starting indentation of the sectin
)
