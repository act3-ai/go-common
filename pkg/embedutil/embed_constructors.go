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
func NewCategory(key, title, manpagePrefix string, manpageExt int8, docs ...*Document) *Category {
	cat := &Category{
		Key:   key,
		Title: title,
		Docs:  docs,
	}

	// Set manpage extensions
	for _, doc := range cat.Docs {
		doc.manpagePrefix = manpagePrefix
		doc.manpageExt = manpageExt
	}

	return cat
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

// LoadMarkdownString loads a markdown string into a Document
func LoadMarkdownString(key, title, name string, data string) *Document {
	d := &Document{
		Key:      key,
		Title:    title,
		name:     name,
		Contents: []byte(data),
		encoding: EncodingMarkdown,
	}
	return d
}

// LoadMarkdownBytes loads markdown bytes into a Document
func LoadMarkdownBytes(key, title, name string, data []byte) *Document {
	d := &Document{
		Key:      key,
		Title:    title,
		name:     name,
		Contents: data,
		encoding: EncodingMarkdown,
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

// LoadJSONSchemaString loads a JSON Schema definition string into a Document
func LoadJSONSchemaString(key, title, name, data string) *Document {
	d := &Document{
		Key:        key,
		Title:      title,
		name:       name,
		Contents:   []byte(data),
		manpageExt: 5,
		encoding:   EncodingJSONSchema,
	}
	return d
}

// LoadJSONSchemaBytes loads JSON Schema definition bytes into a Document
func LoadJSONSchemaBytes(key, title, name string, data []byte) *Document {
	d := &Document{
		Key:        key,
		Title:      title,
		name:       name,
		Contents:   data,
		manpageExt: 5,
		encoding:   EncodingJSONSchema,
	}
	return d
}
