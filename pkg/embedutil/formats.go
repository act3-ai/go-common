package embedutil

import (
	"github.com/cpuguy83/go-md2man/v2/md2man"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Format represents the output format for embedded documents
type Format string

const (
	Markdown Format = "md"   // Markdown represents Markdown output
	HTML     Format = "html" // HTML represents HTML output
	Manpage  Format = "man"  // Manpage represents manpage output
)

// indexable checks if the output format is indexable
func (f Format) indexable() bool {
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

// formatManpage converts a markdown document to a roff format manpage
func formatManpage(data []byte) ([]byte, error) {
	return md2man.Render(data), nil
}

var htmlOpts = &copyOpts{
	PathFunc: func(path string) (string, error) {
		// Convert file extension to html
		return setExtension(path, "html"), nil
	},
	ContentFunc: formatHTML,
}

// formatHTML converts a markdown document to HTML
func formatHTML(data []byte) ([]byte, error) {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(data)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	out := markdown.Render(doc, renderer)

	// // Add simple styling
	// out = append([]byte(heredoc.Doc(`
	// 	<head>
	// 	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.min.css">
	// 	</head>
	// `)), out...)

	return out, nil
}

// represents a conversion from encoding format to output format
type conversion struct {
	Encoding
	Format
}

type conversionFunc func(data []byte) ([]byte, error)

var (
	noopConversion = func(data []byte) ([]byte, error) {
		return data, nil
	}

	// Maps an input and output format to a conversion function
	supportedConversions = map[conversion]conversionFunc{
		{EncodingMarkdown, Markdown}:   noopConversion,
		{EncodingMarkdown, Manpage}:    formatManpage,
		{EncodingMarkdown, HTML}:       formatHTML,
		{EncodingJSONSchema, Markdown}: noopConversion,
		{EncodingJSONSchema, Manpage}:  noopConversion,
		{EncodingJSONSchema, HTML}:     noopConversion,
		// {EncodingCRD, Markdown}:        noopConversion,
		// {EncodingCRD, Manpage}:         noopConversion,
		// {EncodingCRD, HTML}:            noopConversion,
	}
)
