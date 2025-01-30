package termdoc

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AdditionalHelpTopic creates a cobra command that will be surfaced as an "Additional Help Topic".
//
// When run, the content will be formatted by the Formatter.
func AdditionalHelpTopic(name, short, markdownContent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  markdownContent,
		Args:  cobra.ExactArgs(0),
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Println(AutoColorFormat().Format(cmd.Long))
	})
	return cmd
}
