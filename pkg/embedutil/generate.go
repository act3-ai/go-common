package embedutil

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"git.act3-ace.com/ace/go-common/pkg/embedutil/dumpfs"
	"git.act3-ace.com/ace/go-common/pkg/fsutil"
)

// Options stores configuration for rendering embedded documentation
type Options struct {
	Format Format    // Output format
	Types  []DocType // Documentation types to generate
	Index  bool      // Generate a documentation index file (format-dependent)
	Flat   bool      // Generate documentation in a flat directory structure
}

// Write outputs all embedded documentation in the outputDir
func (docs *Documentation) Write(outputDir string, opts *Options) error {
	outFS, err := newFS(outputDir)
	if err != nil {
		return err
	}

	if opts.TypeRequested(TypeCommands) && docs.Command != nil {
		cmdDir := outputDir
		if !opts.Flat {
			cmdDir = filepath.Join(outputDir, "cli")
		}

		// Create FS for the category's docs
		cmdFS, err := newFS(cmdDir)
		if err != nil {
			return err
		}

		// Generate CLI documentation
		err = renderCommandDocs(docs.Command, cmdFS, opts)
		if err != nil {
			return err
		}
	}

	if opts.TypeRequested(TypeGeneral) {
		// Generate each category
		for _, cat := range docs.Categories {
			catDir := outputDir
			if !opts.Flat {
				catDir = filepath.Join(outputDir, cat.DirName())
			}

			// Create FS for the category's docs
			catFS, err := newFS(catDir)
			if err != nil {
				return err
			}

			for _, doc := range cat.Docs {
				err = doc.write(catFS, opts)
				if err != nil {
					return err
				}
			}
		}
	}

	_, err = fmt.Println("Generated documentation: " + outputDir)
	if err != nil {
		return err
	}

	// Check if we can index the output format and if it was requested
	if opts.Format.Indexable() && opts.Index {
		// Create an index file (either README.md or index.html)
		err = docs.writeIndex(outFS, opts)
		if err != nil {
			return err
		}

		_, err = fmt.Println("Generated documentation index: " + filepath.Join(outputDir, opts.Format.IndexFile()))
		if err != nil {
			return err
		}

		// if opts.Format == HTML {
		// 	absIndex, err := filepath.Abs(filepath.Join(outputDir, opts.Format.IndexFile()))
		// 	if err != nil {
		// 		return err
		// 	}

		// 	fmt.Println("Open documentation in your browser: file://" + absIndex)
		// }
	}

	return nil
}

// Render command documentation into the specified format
func renderCommandDocs(cmd *cobra.Command, outFS *fsutil.FSUtil, opts *Options) error {
	cmd.DisableAutoGenTag = true // disable the cobra-generated footer

	switch opts.Format {
	case Manpage:
		// Generate manpages from the commands
		err := doc.GenManTree(cmd, nil, outFS.RootDir)
		if err != nil {
			return fmt.Errorf("failed to document commands: %w", err)
		}
	case Markdown:
		err := doc.GenMarkdownTree(cmd, outFS.RootDir)
		if err != nil {
			return fmt.Errorf("failed to document commands: %w", err)
		}
	case HTML:
		// Create temp directory to write markdown documentation into
		tempFS, err := newTempFS(cmd.Name() + "-command-docs")
		if err != nil {
			return err
		}

		// Generate markdown docs into temp directory
		err = doc.GenMarkdownTree(cmd, tempFS.RootDir)
		if err != nil {
			return fmt.Errorf("failed to document commands: %w", err)
		}

		// Dump markdown files from temp directory to destination,
		// converting files to HTML along the way
		_, err = dumpfs.DumpFS(tempFS, outFS, htmlOpts)
		if err != nil {
			return err
		}

		// Clean up temp directory
		err = tempFS.Close()
		if err != nil {
			return err
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
