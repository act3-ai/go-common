package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/embedutil"
)

// NewInfoCmd creates an info command that allows the viewing of embedded documentation
// in the terminal, converted to Markdown and formatted with glamour
func NewInfoCmd(docs *embedutil.Documentation) *cobra.Command {
	var infoCmd = &cobra.Command{
		Use:   "info <topic>",
		Short: "View detailed documentation for the tool",
		Long:  "The info command provides detailed documentation in your terminal.",
	}

	// Add subcommands for each provided document
	for _, cat := range docs.Categories {

		// Add a command group for the category
		infoCmd.AddGroup(&cobra.Group{
			ID:    cat.Key,
			Title: cat.Title,
		})

		// Add subcommands for each document in the category
		for _, doc := range cat.Docs {
			subCmd := newDocCmd(doc)

			// Associate command with the category's command group
			subCmd.GroupID = cat.Key

			// Add the command the root info command
			infoCmd.AddCommand(subCmd)
		}
	}

	return infoCmd
}

// Creates a command to render a single document in the terminal
func newDocCmd(doc *embedutil.Document) *cobra.Command {
	var writeDir string

	cmd := &cobra.Command{
		Use:   doc.Key,
		Short: doc.Title,
		Long:  fmt.Sprintf("View the %q document in your terminal.", doc.Title),
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			contents, err := doc.Render(embedutil.Markdown)
			if err != nil {
				return err
			}

			if writeDir != "" {
				if err := os.MkdirAll(writeDir, 0775); err != nil {
					return err
				}

				file := doc.RenderedName(embedutil.Markdown)

				err = os.WriteFile(filepath.Join(writeDir, file), contents, 0644)
				if err != nil {
					return err
				}

				cmd.Printf("Wrote the %q document: %s\n", doc.Title, filepath.Join(writeDir, file))
				return nil
			}

			r, _ := glamour.NewTermRenderer(
				// detect background color and pick either the default dark or light theme
				glamour.WithAutoStyle(),
			)

			// Show help for docs
			rendered, err := r.RenderBytes(contents)
			if err != nil {
				return fmt.Errorf("could not format contents: %w", err)
			}

			cmd.Println(string(rendered))
			return nil
		},
	}

	cmd.Flags().StringVarP(&writeDir, "write", "w", "", "write the document to a Markdown file (optionally specify a target directory)")
	cmd.Flags().Lookup("write").NoOptDefVal = "."

	return cmd
}
