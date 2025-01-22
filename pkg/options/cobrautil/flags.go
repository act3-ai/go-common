// Package cobrautil defines utility wrapper functions for common cobra flag handling tasks.
package cobrautil

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MarkFlagRequired instructs the various shell completion implementations to
// prioritize the named flag when performing completion,
// and causes your command to report an error if invoked without the flag.
//
// Wrapper for cobra.MarkFlagRequired accepting a flag object instead of its name.
func MarkFlagRequired(f *pflag.Flag) {
	s := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	s.AddFlag(f)
	err := cobra.MarkFlagRequired(s, f.Name)
	if err != nil {
		panic(fmt.Errorf("marking flag required: %w", err))
	}
	// Actual implementation:
	// flagutil.SetAnnotation(f, cobra.BashCompOneRequiredFlag, "true")
}

// MarkFlagFilename instructs the various shell completion implementations to
// limit completions for the named flag to the specified file extensions.
//
// Wrapper for cobra.MarkFlagFilename accepting a flag object instead of its name.
func MarkFlagFilename(f *pflag.Flag, extensions ...string) {
	s := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	s.AddFlag(f)
	err := cobra.MarkFlagFilename(s, f.Name, extensions...)
	if err != nil {
		panic(fmt.Errorf("marking flag filename: %w", err))
	}
	// Actual implementation:
	// flagutil.SetAnnotation(f, cobra.BashCompFilenameExt, extensions...)
}

// MarkFlagCustom adds the BashCompCustom annotation to the named flag, if it exists.
// The bash completion script will call the bash function f for the flag.
//
// This will only work for bash completion.
// It is recommended to instead use c.RegisterFlagCompletionFunc(...) which allows
// to register a Go function which will work across all shells.
//
// Wrapper for cobra.MarkFlagCustom accepting a flag object instead of its name.
func MarkFlagCustom(f *pflag.Flag, bashFunction string) {
	s := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	s.AddFlag(f)
	err := cobra.MarkFlagCustom(s, f.Name, bashFunction)
	if err != nil {
		panic(fmt.Errorf("marking flag custom: %w", err))
	}
	// Actual implementation:
	// flagutil.SetAnnotation(f, cobra.BashCompCustom, bashFunction)
}

// MarkFlagDirname instructs the various shell completion implementations to
// limit completions for the named flag to directory names.
//
// Wrapper for cobra.MarkFlagDirname accepting a flag object instead of its name.
func MarkFlagDirname(f *pflag.Flag) {
	s := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	s.AddFlag(f)
	err := cobra.MarkFlagDirname(s, f.Name)
	if err != nil {
		panic(fmt.Errorf("marking flag dirname: %w", err))
	}
	// Actual implementation:
	// flagutil.SetAnnotation(f, cobra.BashCompSubdirsInDir)
}

// MarkFlagsRequiredTogether marks the given flags with annotations so that Cobra errors
// if the command is invoked with a subset (but not all) of the given flags.
//
// Wrapper for cobra.Command.MarkFlagsRequiredTogether accepting flag objects instead of names.
func MarkFlagsRequiredTogether(cmd *cobra.Command, flags ...*pflag.Flag) {
	names := make([]string, 0, len(flags))
	for _, f := range flags {
		names = append(names, f.Name)
	}
	cmd.MarkFlagsRequiredTogether(names...)
}

// MarkFlagsOneRequired marks the given flags with annotations so that Cobra errors
// if the command is invoked without at least one flag from the given set of flags.
//
// Wrapper for cobra.Command.MarkFlagsOneRequired accepting flag objects instead of names.
func MarkFlagsOneRequired(cmd *cobra.Command, flags ...*pflag.Flag) {
	names := make([]string, 0, len(flags))
	for _, f := range flags {
		names = append(names, f.Name)
	}
	cmd.MarkFlagsOneRequired(names...)
}

// MarkFlagsMutuallyExclusive marks the given flags with annotations so that Cobra errors
// if the command is invoked with more than one flag from the given set of flags.
//
// Wrapper for cobra.Command.MarkFlagsOneRequired accepting flag objects instead of names.
func MarkFlagsMutuallyExclusive(cmd *cobra.Command, flags ...*pflag.Flag) {
	names := make([]string, 0, len(flags))
	for _, f := range flags {
		names = append(names, f.Name)
	}
	cmd.MarkFlagsMutuallyExclusive(names...)
}

// FlagCompletionFunc defines a shell completion function for a flag, used by cobra commands.
type FlagCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// RegisterFlagCompletionFunc should be called to register a function to provide completion for a flag.
//
// Wrapper for cobra.Command.RegisterFlagCompletionFunc accepting a flag object instead of its name.
func RegisterFlagCompletionFunc(cmd *cobra.Command, flag *pflag.Flag, f FlagCompletionFunc) error {
	return cmd.RegisterFlagCompletionFunc(flag.Name, f) //nolint:wrapcheck
}

// GetFlagCompletionFunc returns the completion function for the given flag of the command, if available.
//
// Wrapper for cobra.Command.GetFlagCompletionFunc accepting a flag object instead of its name.
func GetFlagCompletionFunc(cmd *cobra.Command, flag *pflag.Flag) (FlagCompletionFunc, bool) {
	return cmd.GetFlagCompletionFunc(flag.Name)
}
