package termdoc

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/act3-ai/asce/go-common/pkg/termdoc/mdfmt"
)

// AdditionalHelpTopic creates a cobra command that will be surfaced as an "Additional Help Topic".
//
// When run, the content will be formatted by the Formatter.
func AdditionalHelpTopic(name, short string, markdownContent string, format *mdfmt.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  markdownContent,
		Args:  cobra.ExactArgs(0),
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		out := cmd.OutOrStdout()
		_, err := fmt.Fprintln(out, format.Format(cmd.Long))
		if err != nil {
			cmd.PrintErrln(cmd.ErrPrefix() + err.Error())
		}
	})
	return cmd
}

// lazyLongMessageAnno is the annotation set on commands whose "long" message is produced lazily.
const lazyLongMessageAnno = "termdoc_lazy_long_message"

// HasLazyLongMessage reports whether cmd's "long" message is produced lazily.
func HasLazyLongMessage(cmd *cobra.Command) bool {
	return cmd.Annotations != nil && cmd.Annotations[lazyLongMessageAnno] == "true"
}

// LazyAdditionalHelpTopic creates a cobra command that will be surfaced as an "Additional Help Topic".
//
// The content is produced by the contentFunc when the command is called.
//
// When run, the content will be formatted by the Formatter.
func LazyAdditionalHelpTopic(name, short string, contentFunc func(cmd *cobra.Command, args []string) (string, error), format *mdfmt.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.ExactArgs(0),
		Annotations: map[string]string{
			lazyLongMessageAnno: "true",
		},
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		var errs []error
		markdownContent, err := contentFunc(cmd, args)
		if err != nil {
			errs = append(errs, err)
		}
		if markdownContent != "" {
			cmd.Long = markdownContent
			out := cmd.OutOrStdout()
			_, err = fmt.Fprintln(out, format.Format(markdownContent))
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			for _, err := range errs {
				errStr := cmd.ErrPrefix() + err.Error()
				cmd.PrintErrln(errStr)
				cmd.Long += errStr + "\n"
			}
		}
	})
	return cmd
}
