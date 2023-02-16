package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// NewGendocsCmd is a command to generate the internal CLI documentation in markdown
// additionalManpages is a map of non-generatable man pages to be included (ex. Quick Start Guides, User Guides)
func NewGendocsCmd(additionalManpages embed.FS) *cobra.Command {
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

				return fs.WalkDir(additionalManpages, ".", func(path string, d fs.DirEntry, err error) error { //nolint:wrapcheck
					if err != nil {
						return err
					}

					if d.IsDir() {
						return os.MkdirAll(filepath.Join(docsPath, path), 0777) //nolint:wrapcheck
					}

					content, err := additionalManpages.ReadFile(path)
					if err != nil {
						return err //nolint:wrapcheck
					}

					return os.WriteFile(filepath.Join(docsPath, path), content, 0644) //nolint:wrapcheck
				})
			}

			return fmt.Errorf("incorrect value for format")
		},
	}
	gendocsCmd.Flags().StringVarP(&format, "format", "f", "md", "Set output documentation format. Supports \"md\" for markdown or \"man\" for manpage")
	return gendocsCmd
}
