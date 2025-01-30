package mdfmt

// MDLocation describes the current location of text in a markdown document.
type MDLocation struct {
	Level          int  // Header level of the current section
	Header         bool // Line is a header line
	CodeBlock      bool // Line is within a multiline code block
	CodeBlockLevel int  // Number of "`" characters used to start the multiline code block
	Comment        bool // Line is in an HTML comment
}

// Formatter formats Markdown for terminal output.
type Formatter struct {
	Header  func(text string, loc MDLocation) string      // reformats headers
	Link    func(text, url string, loc MDLocation) string // reformats links
	Code    func(code string, loc MDLocation) string      // reformats inline code blocks
	Bold    func(text string, loc MDLocation) string      // reformats bolded text
	Italics func(text string, loc MDLocation) string      // reformats italicized text
	Indent  func(loc MDLocation) string                   // produces indent for a line's location

	// produce column width for wrapping
	// (nil function or 0 return value disables wrapping)
	Columns func() int
}

// StaticColumns is a static columns setting.
func StaticColumns(cols int) func() int {
	return func() int {
		return cols
	}
}
