package embedutil

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"git.act3-ace.com/ace/go-common/pkg/fsutil"
)

// DocType represents an overarching area of documentation
type DocType string

const (
	TypeAll      DocType = "all"      // TypeAll represents all types
	TypeGeneral  DocType = "general"  // TypeGeneral represents general documentation
	TypeCommands DocType = "commands" // TypeCommands represents CLI command documentation
	TypeSchemas  DocType = "schemas"  // TypeSchemas represents API schema documentation
)

// TypeRequested checks if a type was requested from the options
func (opts *Options) TypeRequested(checkType DocType) bool {
	for _, requested := range opts.Types {
		if requested == checkType || requested == TypeAll {
			return true
		}
	}
	return false
}

// RenderedName produces the output file name of a document based on the format
func (doc *Document) RenderedName(format Format) string {
	// Output JSON Schema docs as-is
	if doc.encoding == EncodingJSONSchema {
		return doc.name
	}

	// Set manpage extensions
	if format == Manpage {
		return setExtension(doc.Key, doc.ManpageExt())
	}

	// Set extensions for MD or HTML
	return setExtension(doc.name, string(format))
}

// ManpageExt produces the extension to be used for this document when represented as a manpage
func (doc *Document) ManpageExt() string {
	if doc.manpageExt != 0 {
		// Use specified number
		return strconv.Itoa(int(doc.manpageExt))
	}

	// Set schemas to 5 for config docs
	if doc.encoding == EncodingJSONSchema {
		return "5"
	}

	// Default to 1
	return "1"
}

// Render produces the document's content in the requested format
func (doc *Document) Render(format Format) ([]byte, error) {
	// Output JSON Schema docs as-is
	if doc.encoding == EncodingJSONSchema {
		return doc.Contents, nil
	}

	noConvErr := fmt.Errorf("unsupported conversion: cannot convert %q from %s to %s", doc.name, doc.encoding, format)

	// Render docs to specified output format
	switch format {
	case Manpage:
		switch doc.encoding {
		case EncodingManpage, EncodingJSONSchema, EncodingRaw:
			return doc.Contents, nil
		case EncodingMarkdown:
			return FormatManpage(doc.Contents)
		default:
			return nil, noConvErr
		}
	case Markdown:
		switch doc.encoding {
		case EncodingMarkdown, EncodingJSONSchema, EncodingRaw:
			return doc.Contents, nil
		default:
			return nil, noConvErr
		}
	case HTML:
		switch doc.encoding {
		case EncodingHTML, EncodingJSONSchema, EncodingRaw:
			return doc.Contents, nil
		case EncodingMarkdown:
			return FormatHTML(doc.Contents)
		default:
			return nil, noConvErr
		}
	default:
		return nil, noConvErr
	}
}

// helper to write a doc to an FS
func (doc *Document) write(outFS *fsutil.FSUtil, opts *Options) error {
	// Evaluate output path
	path := doc.RenderedName(opts.Format)

	// Evaluate document contents
	contents, err := doc.Render(opts.Format)
	if err != nil {
		return err
	}

	// Write the file to outFS
	return outFS.AddFileWithData(path, contents)
}

// Replaces the current file extension of path with newExtension
func setExtension(path, newExtension string) string {
	return removeExtension(path) + "." + strings.TrimPrefix(newExtension, ".")
}

// Remove extension from the path
func removeExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
