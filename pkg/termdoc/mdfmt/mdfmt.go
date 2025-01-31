package mdfmt

// Location describes the current location of text in a markdown document.
type Location struct {
	Level          int    // Header level of the current section
	Header         bool   // Line is a header line
	CodeBlock      bool   // Line is within a multiline code block
	CodeBlockLang  string // Language identifier for the code block
	CodeBlockLevel int    // Number of "`" characters used to start the multiline code block
	Comment        bool   // Line is in an HTML comment
}

// Formatter formats Markdown for terminal output.
type Formatter struct {
	Header    func(text string, loc Location) string      // reformats headers
	Link      func(text, url string, loc Location) string // reformats links
	Code      func(code string, loc Location) string      // reformats inline code blocks
	CodeBlock func(code string, loc Location) string      // reformats multiline code blocks
	Bold      func(text string, loc Location) string      // reformats bolded text
	Italics   func(text string, loc Location) string      // reformats italicized text
	Indent    func(loc Location) string                   // produces indent for a line's location

	// produce column width for wrapping
	// (nil function or 0 return value disables wrapping)
	Columns func() int

	// CodeBlockWrapMode signifies a code block wrapping style.
	CodeBlockWrapMode WrapMode
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
