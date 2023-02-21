package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// NewGendocsCmd is a command to generate the internal CLI documentation in markdown
// additionalManpages is a map of non-generatable man pages to be included (ex. Quick Start Guides, User Guides)
func NewGendocsCmd(additionalManpages fs.FS) *cobra.Command {
	var format string
	var gendocsCmd = &cobra.Command{
		Use:    "gendocs <docs location>",
		Short:  "Generate documentation from usage descriptions",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			docsPath := args[0]

			if format == "md" {
				return doc.GenMarkdownTree(cmd.Root(), docsPath) //nolint:wrapcheck
			} else if format == "man" {
				err := doc.GenManTree(cmd.Root(), nil, docsPath)
				if err != nil {
					return err //nolint:wrapcheck
				}

				// Map filename to src path
				srcMap := map[string]string{}

				return fs.WalkDir(additionalManpages, ".", func(path string, d fs.DirEntry, err error) error { //nolint:wrapcheck
					if err != nil {
						return err
					}

					if d.IsDir() {
						return nil
					}

					src, err := additionalManpages.Open(path)
					if err != nil {
						return fmt.Errorf("could not open manpage %q: %w", path, err)
					}

					// Flatten to just filename
					filename := filepath.Base(path)

					// Check if filename has already been found in the fs.FS so files are not overwritten
					if srcPath, ok := srcMap[filename]; ok {
						return fmt.Errorf("duplicate filename %q from %q and %q", filename, path, srcPath)
					}

					srcMap[filename] = path

					destPath := filepath.Join(docsPath, filename)
					dst, err := os.Create(destPath)
					if err != nil {
						return fmt.Errorf("could not create file %q: %w", destPath, err)
					}

					_, err = io.Copy(dst, src)
					return fmt.Errorf("could not copy content to %q: %w", dst.Name(), err)
				})
			}

			return fmt.Errorf("incorrect value for format")
		},
	}
	gendocsCmd.Flags().StringVarP(&format, "format", "f", "md", "Set output documentation format. Supports \"md\" for markdown or \"man\" for manpage")
	return gendocsCmd
}
