package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/act3-ai/go-common/pkg/version"
)

// versionOptions is the options for the version
type versionOptions struct {
	version.Info
	Short bool
}

// Run is the action method
func (action *versionOptions) Run(out io.Writer) error {
	if action.Short {
		_, err := fmt.Fprintln(out, action.Version)
		return err
	}
	_, err := fmt.Fprintf(out, "%#v\n", action.Info)
	return err
}

// NewVersionCmd creates a new "version" subcommand
func NewVersionCmd(info version.Info) *cobra.Command {
	options := &versionOptions{
		Info: info,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.Run(cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVarP(&options.Short, "short", "s", false, "print just the version (not extra information)")

	return cmd
}
