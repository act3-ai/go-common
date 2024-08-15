package embedutil

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Options stores configuration for rendering embedded documentation
type Options struct {
	Format Format    // Output format
	Types  []DocType // Documentation types to generate
	Index  bool      // Generate a documentation index file (format-dependent)
	Flat   bool      // Generate documentation in a flat directory structure
}

// Write outputs all embedded documentation in the outputDir
func (docs *Documentation) Write(ctx context.Context, outputDir string, opts *Options) error {
	err := os.MkdirAll(outputDir, 0o775)
	if err != nil {
		return fmt.Errorf("writing documentation: %w", err)
	}

	if opts.TypeRequested(TypeCommands) && docs.Command != nil {
		cmdDir := outputDir
		if !opts.Flat && len(opts.Types) > 1 {
			// Create FS for the category's docs
			cmdDir = filepath.Join(outputDir, "cli")
		}

		// Generate CLI documentation
		err = renderCommandDocs(docs.Command, cmdDir, opts)
		if err != nil {
			return err
		}
	}

	if opts.TypeRequested(TypeGeneral) {
		// Generate each category
		for _, cat := range docs.Categories {
			catDir := outputDir
			if !opts.Flat {
				// Create directory for the category's docs
				catDir = filepath.Join(outputDir, cat.dirName())
				err = os.MkdirAll(catDir, 0o775)
				if err != nil {
					return fmt.Errorf("creating document: %w", err)
				}
			}

			for _, doc := range cat.Docs {
				contents, err := doc.Render(opts.Format)
				if err != nil {
					return err
				}

				err = os.WriteFile(filepath.Join(catDir, doc.RenderedName(opts.Format)), contents, 0o644)
				if err != nil {
					return fmt.Errorf("creating document: %w", err)
				}
			}
		}
	}

	slog.InfoContext(ctx, "Generated documentation", slog.String("dir", outputDir), slog.String("format", string(opts.Format)))

	return docs.writeIndex(outputDir, opts)
}

func (docs *Documentation) writeIndex(outputDir string, opts *Options) error {
	// Check if we can index the output format and if it was requested
	if !opts.Format.indexable() || !opts.Index {
		return nil
	}

	// Create an index file (either README.md or index.html)
	index, err := docs.Index(outputDir, opts)
	if err != nil {
		return err
	}

	if index == nil {
		return nil
	}

	indexFile := filepath.Join(outputDir, opts.Format.IndexFile())

	err = os.WriteFile(indexFile, index, 0o644)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}

	_, err = fmt.Println("Generated documentation index: " + indexFile)
	if err != nil {
		return err
	}

	return nil
}

// Render command documentation into the specified format
func renderCommandDocs(cmd *cobra.Command, outputDir string, opts *Options) error {
	cmd.DisableAutoGenTag = true // disable the cobra-generated footer

	switch opts.Format {
	case Manpage:
		// Generate manpages from the commands
		err := doc.GenManTree(cmd, nil, outputDir)
		if err != nil {
			return fmt.Errorf("documenting commands: %w", err)
		}
	case Markdown:
		err := renderMarkdownTree(cmd, outputDir, opts)
		if err != nil {
			return fmt.Errorf("documenting commands: %w", err)
		}
	case HTML:
		tempDir, err := os.MkdirTemp("", cmd.Name()+"-command-docs-*")
		if err != nil {
			return fmt.Errorf("documenting commands: %w", err)
		}

		// Generate markdown docs into temp directory
		err = renderMarkdownTree(cmd, outputDir, opts)
		if err != nil {
			return fmt.Errorf("documenting commands: %w", err)
		}

		// Dump markdown files from temp directory to destination,
		// converting files to HTML along the way
		_, err = copyConvert(tempDir, outputDir, htmlOpts)
		if err != nil {
			return err
		}

		// Clean up temp directory
		if err := os.RemoveAll(tempDir); err != nil {
			return fmt.Errorf("documenting commands: %w", err)
		}
	}
	return nil
}

// // Render documentation for JSON Schema definitions into specified format
// func renderSchemas(schemas fs.FS, outFS *fsutil.FSUtil, opts *GenerateOptions) error {
// 	if opts.Format.Indexed() {
// 		// If writing to an indexable format, render the docs into a subdirectory
// 		var err error
// 		outFS, err = newFS(filepath.Join(outFS.RootDir, "schemas"))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	_, err := dumpFS(schemas, outFS, nil)
// 	return err
// }
