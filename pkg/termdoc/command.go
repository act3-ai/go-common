package termdoc

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/mdfmt"
)

// AdditionalHelpTopic creates a cobra command that will be surfaced as an "Additional Help Topic".
//
// When run, the content will be formatted by the Formatter.
func AdditionalHelpTopic(name, short, markdownContent string, format *mdfmt.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  markdownContent,
		Args:  cobra.ExactArgs(0),
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Println(format.Format(cmd.Long))
	})
	return cmd
}
