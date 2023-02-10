package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// GendocsOptions stores options for the gendocs command
type GendocsOptions struct {
	Format             string            // Documentation format, either "md" or "man"
	AdditionalManpages map[string][]byte // Non-generatable man pages to be included (ex. Quick Start Guides, User Guides)
}

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

// Run runs the Gendocs action
func (action *GendocsOptions) Run(ctx context.Context, cmd *cobra.Command, docsPath string) error {
	if action.Format == "md" {
		return doc.GenMarkdownTree(cmd.Root(), docsPath) //nolint:wrapcheck
	} else if action.Format == "man" {
		err := doc.GenManTree(cmd.Root(), nil, docsPath)
		if err != nil {
			return err //nolint:wrapcheck
		}

		for name, content := range action.AdditionalManpages {
			if err := os.WriteFile(name, content, 0666); err != nil {
				return err //nolint:wrapcheck
			}
		}
	}

	return fmt.Errorf("incorrect value for format")
}
