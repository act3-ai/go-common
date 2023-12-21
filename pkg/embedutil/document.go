package embedutil

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
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
	conv := conversion{doc.encoding, format}
	convFunc, ok := supportedConversions[conv]
	if !ok {
		return nil, fmt.Errorf("unsupported conversion: cannot convert %q from %s to %s", doc.name, doc.encoding, format)
	}

	return convFunc(doc.Contents)
}

// Replaces the current file extension of path with newExtension
func setExtension(path, newExtension string) string {
	return removeExtension(path) + "." + strings.TrimPrefix(newExtension, ".")
}

// Remove extension from the path
func removeExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
