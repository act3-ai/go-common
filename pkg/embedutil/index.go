package embedutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
)

// Index creates an index file for the documentation in the requested format
func (docs *Documentation) Index(outputDir string, opts *Options) ([]byte, error) {
	if !opts.Format.indexable() || !opts.Index {
		// Skip indexing if not enabled in options
		// or not supported by the output format (manpages)
		return nil, nil
	}

	// Generate a markdown-formatted index file
	index, err := docs.generateMarkdownIndex(outputDir, opts)
	if err != nil {
		return index, err
	}

	switch opts.Format {
	case HTML:
		index, err = formatHTML(index)
		if err != nil {
			return index, err
		}
		return index, nil
	case Markdown:
		return index, err
	default:
		return nil, nil
	}
}

func (docs *Documentation) generateMarkdownIndex(outputDir string, opts *Options) ([]byte, error) {
	index := new(bytes.Buffer)

	_, _ = fmt.Fprint(index, heredoc.Docf(`
		# %s

		Documentation for %s is organized as follows:

	`, docs.Title, docs.Command.Name()))

	groupNameTemplate := "\n## %s\n\n"
	mdLinkTemplate := "- [%s](./%s)\n"

	addCategory := func(cat *Category) error {
		if len(cat.Docs) == 0 {
			return nil
		}

		// Append section header
		_, _ = fmt.Fprintf(index, groupNameTemplate, cat.Title)

		for _, doc := range cat.Docs {
			docPath := doc.RenderedName(opts.Format)
			if !opts.Flat {
				// Append category dir for non-flat renders
				docPath = filepath.Join(cat.dirName(), docPath)
			}
			// Append file link
			_, _ = fmt.Fprintf(index, mdLinkTemplate, doc.Title, docPath)
		}

		return nil
	}

	outputFS := os.DirFS(outputDir)

	addGroupFromDir := func(groupName string, dir string) error {
		entries, err := fs.ReadDir(outputFS, dir)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}

		if len(entries) == 0 {
			return nil
		}

		// Append section header
		_, _ = fmt.Fprintf(index, groupNameTemplate, groupName)

		for _, entry := range entries {
			// Skip directories
			if entry.IsDir() {
				continue
			}

			// Append link to file
			file := entry.Name()
			// Append file link
			_, _ = fmt.Fprintf(index, mdLinkTemplate, removeExtension(file), filepath.Join(dir, file))
		}

		return nil
	}

	// Index each category
	for _, cat := range docs.Categories {
		err := addCategory(cat)
		if err != nil {
			return nil, err
		}
	}

	// Index CLI documentation
	err := addGroupFromDir("CLI Commands", "cli")
	if err != nil {
		return nil, err
	}

	return index.Bytes(), nil
}
