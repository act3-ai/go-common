package embedutil

import (
	"github.com/cpuguy83/go-md2man/v2/md2man"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"git.act3-ace.com/ace/go-common/pkg/embedutil/dumpfs"
)

// Format represents the output format for embedded documents
type Format string

const (
	Markdown Format = "md"   // Markdown represents Markdown output
	HTML     Format = "html" // HTML represents HTML output
	Manpage  Format = "man"  // Manpage represents manpage output
)

// Indexable checks if the output format is indexable
func (f Format) Indexable() bool {
	return f == Markdown || f == HTML
}

// IndexFile returns the index file name corresponding to the format
func (f Format) IndexFile() string {
	switch f {
	case Markdown:
		return "README.md"
	case HTML:
		return "index.html"
	default:
		return ""
	}
}

// FormatManpage converts a markdown document to a roff format manpage
func FormatManpage(data []byte) ([]byte, error) {
	return md2man.Render(data), nil
}

var htmlOpts = &dumpfs.Options{
	PathFunc: func(path string) (string, error) {
		// Convert file extension to html
		return setExtension(path, "html"), nil
	},
	ContentFunc: FormatHTML,
}

// FormatHTML converts a markdown document to HTML
func FormatHTML(data []byte) ([]byte, error) {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(data)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer), nil
}
