package cobrautil

import "github.com/spf13/cobra"

// WalkCommands calls f for the root command and all subcommands recursively.
func WalkCommands(root *cobra.Command, f func(cmd *cobra.Command)) {
	f(root)
	for _, child := range root.Commands() {
		WalkCommands(child, f)
	}
}
