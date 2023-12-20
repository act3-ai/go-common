package embedutil

import (
	"io/fs"
	"path/filepath"
)

// Encoding represents an embedded document's encoding
type Encoding string

const (
	EncodingMarkdown   Encoding = "md"         // EncodingMarkdown represents a Markdown-encoded document
	EncodingManpage    Encoding = "man"        // EncodingManpage represents a manpage document
	EncodingJSONSchema Encoding = "jsonschema" // EncodingJSONSchema represents a JSON-encoded JSON Schema definition
	EncodingCRD        Encoding = "crd"        // EncodingCRD represents a YAML-encoded CustomResourceDefinition
	EncodingHTML       Encoding = "html"       // EncodingHTML represents an HTML document
	EncodingRaw        Encoding = "raw"        // EncodingRaw represents a raw document
)

// NewCategory initializes a new Category object
func NewCategory(key, title string, docs ...*Document) *Category {
	return &Category{
		Key:   key,
		Title: title,
		Docs:  docs,
	}
}

// LoadMarkdown loads a markdown file into a Document
// name must be the path to the document in filesys
func LoadMarkdown(key, title, name string, filesys fs.FS) *Document {
	d := &Document{
		Key:      key,
		Title:    title,
		name:     filepath.Base(name),
		encoding: EncodingMarkdown,
	}

	var err error
	d.Contents, err = fs.ReadFile(filesys, name)
	if err != nil {
		panic(err)
	}
	return d
}

// LoadJSONSchema loads a JSON Schema definition into a Document
// name must be the path to the document in filesys
func LoadJSONSchema(key, title, name string, filesys fs.FS) *Document {
	d := &Document{
		Key:        key,
		Title:      title,
		name:       filepath.Base(name),
		manpageExt: 5,
		encoding:   EncodingJSONSchema,
	}

	var err error
	d.Contents, err = fs.ReadFile(filesys, name)
	if err != nil {
		panic(err)
	}
	return d
}
