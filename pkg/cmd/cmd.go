package cmd

import "github.com/spf13/cobra"

// AddGroupedCommands is a helper function for adding command groups
// All commands will be added to root and associated with group
func AddGroupedCommands(root *cobra.Command, group *cobra.Group, cmds ...*cobra.Command) {
	// Add the group to the root command
	root.AddGroup(group)

	for _, cmd := range cmds {
		// Set command group ID to the provided group
		cmd.GroupID = group.ID

		// Add the command the the root command
		root.AddCommand(cmd)
	}
}
