package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// NewGendocsCmd is a command to generate the internal CLI documentation in markdown
func NewGendocsCmd() *cobra.Command {
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
				return doc.GenManTree(cmd.Root(), nil, docsPath) //nolint:wrapcheck
			}
			return fmt.Errorf("incorrect value for format")
		},
	}
	gendocsCmd.Flags().StringVarP(&format, "format", "f", "md", "Set output documentation format. Supports \"md\" for markdown or \"man\" for manpage")
	return gendocsCmd
}
