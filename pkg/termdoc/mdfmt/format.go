package mdfmt

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Markdown component regexes
var (
	mdBoldUndRegex   = regexp.MustCompile(wordUnd(`__([^_]+)__`))                   // __bold__
	mdBoldAstRegex   = regexp.MustCompile(wordAst(`\*\*([^\*]+)\*\*`))              // **bold**
	mdItalicUndRegex = regexp.MustCompile(wordUnd(`_([^_]+)_`))                     // _italic_
	mdItalicAstRegex = regexp.MustCompile(wordAst(`\*([^\*]+)\*`))                  // *italic*
	mdCodeRegex      = regexp.MustCompile("`[^`]+`")                                // `code`
	mdLinkRegex      = regexp.MustCompile(`\[(?P<text>[^\]]+)\]\((?<url>[^\)]+)\)`) // [text](url)
)

const (
	codeBlockStart = "```"
	commentStart   = "<!--"
	commentEnd     = "-->"
	tableStart     = "|"
)

func wordAst(re string) string {
	return `(?P<before>^|[^\w\*])` + re + `(?P<after>[^\w\*]|$)`
}

func wordUnd(re string) string {
	return `(?P<before>^|[^\w_])` + re + `(?P<after>[^\w_]|$)`
}

// Format formats markdown text according the Formatter's rules.
func (format *Formatter) Format(markdownText string) string {
	cols := 0
	if format.Columns != nil {
		cols = format.Columns()
	}

	lines := strings.Split(markdownText, "\n")
	formatted := make([]string, 0, len(lines))
	var loc Location
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
				loc.CodeBlockLang = ""
				codeBlockStop = ""
			} else if format.CodeBlock != nil {
				line = format.CodeBlock(line, loc)
			}
		// Start code block
		case strings.HasPrefix(lineTrimSpace, codeBlockStart):
			loc.CodeBlock = true
			loc.CodeBlockLevel, loc.CodeBlockLang = parseCodeBlockStart(lineTrimSpace)
			codeBlockStop = strings.Repeat("`", loc.CodeBlockLevel)
			codeBlockIndent = extraIndent(line)
		// In table
		case strings.HasPrefix(lineTrimSpace, tableStart):
			loc.Table = true

			// Assemble list of cells
			var cells []string
			for _, cell := range strings.Split(strings.Trim(lineTrimSpace, "|"), "|") {
				col := len(cells)
				// If there is a previous cell and it ends with the escape character,
				// append the current cell with the escaped pipe character.
				if col > 0 && strings.HasSuffix(cells[col-1], `\`) {
					cells[col-1] += "|" + cell
					continue
				}
				cells = append(cells, cell)
			}

			// Format the text within each cell.
			for i := range cells {
				width := ansi.StringWidth(cells[i])
				cells[i] = format.formatRegularLine(cells[i], loc)
				fmtwidth := ansi.StringWidth(cells[i])
				if width > fmtwidth {
					// Compensate for changed width, if possible
					cells[i] += strings.Repeat(" ", width-fmtwidth)
				}
			}

			// Re-assemble line
			line = "|" + strings.Join(cells, "|") + "|"
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
			// TODO: code block wrapping
			// should it wrap to the level of the starting backticks or to the indentation level
			// of the line within the code block?
			// Wrapping to the indentation level of the line within the code block looks somewhat nicer
			// and is easier to implement.
			var indent string
			switch {
			// Obey code block wrapping mode
			case loc.CodeBlock:
				switch format.CodeBlockWrapMode {
				// Preserve leading whitespace from where the codeblock was started
				case WrapToStartingIndentation:
					indent = codeBlockIndent
				// Preserve leading whitespace in the line
				// Must be determined from the line itself
				default:
					indent = extraIndent(line)
				}
			// Preserve leading whitespace in the line
			// Must be determined from the line itself
			default:
				indent = extraIndent(line)
			}
			// Wrap lines
			line = ansi.Wordwrap(line, cols, " ")
			// Add indent to wrapped lines
			line = strings.ReplaceAll(line, "\n", "\n"+indent)
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

// Performs all non-word-wrap formatting for all markdown lines except those inside multiline code blocks
func (format *Formatter) formatRegularLine(line string, loc Location) string {
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
		line = mdLinkRegex.ReplaceAllStringFunc(line,
			func(s string) string {
				match := mdLinkRegex.FindStringSubmatch(s)
				return format.Link(match[1], match[2], loc)
			})
	}
	if format.Header != nil && loc.Header {
		line = format.Header(strings.TrimSpace(strings.TrimLeft(line, "#")), loc)
	}

	// Markdown inline code formatter:
	codeBlockMatches := mdCodeRegex.FindAllStringIndex(line, -1)
	endPrevious := 0
	linePieces := []string{}
	for _, match := range codeBlockMatches {
		// Add text before code block
		// Get string from end of last code block to start of this one
		before := line[endPrevious:match[0]]
		linePieces = append(linePieces, format.textFormat(before, loc))

		// Add the code block
		codeBlock := line[match[0]:match[1]]
		if format.Code != nil {
			// Add formatted code block
			codeBlockWithoutBackticks := codeBlock[1 : len(codeBlock)-1]
			linePieces = append(linePieces, format.Code(codeBlockWithoutBackticks, loc))
		} else {
			// Add unformatted code block
			linePieces = append(linePieces, codeBlock)
		}

		// Update end index
		endPrevious = match[1]
	}

	// Add formatted remainder (which may be entire line if no code blocks found)
	linePieces = append(linePieces, format.textFormat(line[endPrevious:], loc))

	// Rejoin line
	line = strings.Join(linePieces, "")

	return line
}

func (format *Formatter) textFormat(text string, loc Location) string {
	// Markdown bold formatter:
	if format.Bold != nil {
		// Replace both underline and asterisk notation
		text = mdBoldUndRegex.ReplaceAllStringFunc(text,
			func(s string) string {
				match := mdBoldUndRegex.FindStringSubmatch(s)
				before := match[1]
				inner := format.Bold(match[2], loc)
				after := match[3]
				return before + inner + after
			})
		text = mdBoldAstRegex.ReplaceAllStringFunc(text,
			func(s string) string {
				match := mdBoldAstRegex.FindStringSubmatch(s)
				before := match[1]
				inner := format.Bold(match[2], loc)
				after := match[3]
				return before + inner + after
			})
	}

	// Markdown italic formatter:
	if format.Italics != nil {
		// Replace both underline and asterisk notation
		text = mdItalicUndRegex.ReplaceAllStringFunc(text,
			func(s string) string {
				match := mdItalicUndRegex.FindStringSubmatch(s)
				before := match[1]
				inner := format.Italics(match[2], loc)
				after := match[3]
				return before + inner + after
			})
		text = mdItalicAstRegex.ReplaceAllStringFunc(text,
			func(s string) string {
				match := mdItalicAstRegex.FindStringSubmatch(s)
				before := match[1]
				inner := format.Italics(match[2], loc)
				after := match[3]
				return before + inner + after
			})
	}

	return text
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

func parseCodeBlockStart(s string) (level int, lang string) {
	if strings.HasPrefix(s, "`") {
		level, lang := parseCodeBlockStart(strings.TrimPrefix(s, "`"))
		return 1 + level, lang
	}
	return 0, s
}
