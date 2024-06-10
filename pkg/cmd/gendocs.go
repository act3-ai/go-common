package cmd

import (
	"os"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	embedutil "gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
)

// NewGendocsCmd creates a gendocs command group that allows tools to
// output embedded documentation in various formats
func NewGendocsCmd(docs *embedutil.Documentation) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gendocs",
		Short: "Generate documentation for the tool in various formats",
	}

	cmd.AddCommand(
		newHTMLCmd(docs),
		newMarkdownCmd(docs),
		newManpageCmd(docs),
	)

	return cmd
}

func newHTMLCmd(docs *embedutil.Documentation) *cobra.Command {
	opts := &embedutil.Options{
		Format: embedutil.HTML,
		Types:  []embedutil.DocType{embedutil.TypeGeneral, embedutil.TypeCommands, embedutil.TypeSchemas},
		Index:  true,
		Flat:   false,
	}

	cmd := &cobra.Command{
		Use: "html [dir]",
		Aliases: []string{
			"web",
			"webpage",
		},
		Short: "Generate documentation in HTML format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			os.Setenv("NO_COLOR", "1")
			// disableTermenvColor() // avoid writing ANSI escape codes to files
			return docs.Write(dir, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Index, "index", "i", true, `generate an index.html index file`)
	cmd.Flags().BoolVarP(&opts.Flat, "flat", "f", false, `generate docs in a flat directory structure`)
	// gendocsCmd.Flags().BoolVarP(&opts.Serve, "serve", "s", opts.Serve, "Serve generated docs")

	return cmd
}

func newMarkdownCmd(docs *embedutil.Documentation) *cobra.Command {
	opts := &embedutil.Options{
		Format: embedutil.Markdown,
		Types:  []embedutil.DocType{embedutil.TypeGeneral, embedutil.TypeCommands, embedutil.TypeSchemas},
		Index:  true,
		Flat:   false,
	}

	var onlyCommands bool

	cmd := &cobra.Command{
		Use: "md [dir]",
		Aliases: []string{
			"markdown",
		},
		Short: "Generate documentation in Markdown format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if onlyCommands {
				opts.Types = []embedutil.DocType{embedutil.TypeCommands}
				opts.Index = false
			}

			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			os.Setenv("NO_COLOR", "1")
			// disableTermenvColor() // avoid writing ANSI escape codes to files
			return docs.Write(dir, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Index, "index", "i", true, `generate a README.md index file`)
	cmd.Flags().BoolVarP(&opts.Flat, "flat", "f", false, `generate docs in a flat directory structure`)
	cmd.Flags().BoolVar(&onlyCommands, "only-commands", false, "only generate command documentation")
	cmd.MarkFlagsMutuallyExclusive("only-commands", "index")

	return cmd
}

func newManpageCmd(docs *embedutil.Documentation) *cobra.Command {
	opts := &embedutil.Options{
		Format: embedutil.Manpage,
		Types:  []embedutil.DocType{embedutil.TypeGeneral, embedutil.TypeCommands},
		Index:  false,
		Flat:   true,
	}

	cmd := &cobra.Command{
		Use: "man [dir]",
		Aliases: []string{
			"manpage",
			"manpages",
		},
		Short: "Generate documentation in manpage format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			os.Setenv("NO_COLOR", "1")
			// disableTermenvColor() // avoid writing ANSI escape codes to files
			return docs.Write(dir, opts)
		},
	}

	return cmd
}

// avoid writing ANSI escape codes to files
func disableTermenvColor() {
	termenv.SetDefaultOutput(termenv.NewOutput(termenv.DefaultOutput(), termenv.WithProfile(termenv.Ascii)))
}
