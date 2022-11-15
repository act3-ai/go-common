package action

import (
	"fmt"
	"io"

	"git.act3-ace.com/ace/go-common/pkg/version"
)

// Helm has a good pattern for flags.  https://github.com/helm/helm/blob/main/cmd/helm/version.go

// Version is the action that returns the version
type Version struct {
	version.Info
	Short bool
}

// NewVersion created a new action to output the version
// info is the version to output
func NewVersion(info version.Info) *Version {
	return &Version{
		Info: info,
	}
}

// Run is the action method
func (action *Version) Run(out io.Writer) error {
	if action.Short {
		_, err := fmt.Fprintln(out, action.Version)
		return err
	}
	_, err := fmt.Fprintf(out, "%#v\n", action.Info)
	return err
}
