package embedutil

import (
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

// Documentation configures how different genres of
// embedded documentation will be generated
type Documentation struct {
	// Overall title for the documentation
	Title string

	// Root cobra.Command
	Command *cobra.Command

	// TODO: add Go package docs
	// golang.org/x/tools/cmd/godoc from cs.opensource.google/go/x/tools
	// Pkg      bool

	// Categories stores a list of documentation sub-categories,
	// allowing organization of generated documentation
	// Ordering is obeyed in the indexer
	Categories []*Category
}

// Category is used to group documents
type Category struct {
	Key   string      // Key name for the category in kebab-case
	Title string      // Readable name for the category (can include spaces)
	Docs  []*Document // List of documents contained in the category
}

// DirName produces the directory name used for the category
func (cat *Category) DirName() string {
	if cat.Key == "" {
		cat.Key = strcase.ToKebab(cat.Title)
	}
	return cat.Key
}

// Document represents an embedded document
type Document struct {
	Key        string   // Key name for the file in kebab-case
	Title      string   // Human-readable title for the document
	name       string   // Internal file name
	manpageExt int8     // Manpage extension for the file. Ex: 1 for normal docs, 5 for config docs
	Contents   []byte   // Contents of the document
	encoding   Encoding // Encoding of the file
}

// FindDocument returns the Document with the requested key
func (docs *Documentation) FindDocument(key string) *Document {
	// Show help for docs
	for _, cat := range docs.Categories {
		for _, doc := range cat.Docs {
			if key == doc.Key {
				return doc
			}
		}
	}

	return nil
}

// // DocumentKeys returns a list of all available document keys
// func (docs *Documentation) DocumentKeys() []string {
// 	keys := []string{}

// 	// Show help for docs
// 	for _, cat := range docs.Categories {
// 		for _, doc := range cat.Docs {
// 			keys = append(keys, doc.Key)
// 		}
// 	}

// 	return keys
// }
