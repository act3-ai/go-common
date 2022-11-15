package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/version"
)

// versionOptions is the options for the version
type versionOptions struct {
	version.Info
	Short bool
}

// newVersionOptions create a new version options
// info is the version to output
func newVersionOptions(info version.Info) *versionOptions {
	return &versionOptions{
		Info: info,
	}
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
	action := newVersionOptions(info)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&action.Short, "short", false, "print just the version (not extra information)")

	return cmd
}
