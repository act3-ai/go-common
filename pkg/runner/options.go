package runner

import (
	"context"

	"github.com/spf13/cobra"
)

// RunWithOptions runs a command with options.
func RunWithOptions(ctx context.Context, cmd *cobra.Command, opts ...Option) error {
	WithOptions(cmd)
	return cmd.ExecuteContext(ctx)
}

// WithOptions modifies a command with options.
func WithOptions(cmd *cobra.Command, opts ...Option) {
	for _, opt := range opts {
		opt(cmd)
	}
}

// Option defines an option for running a command.
type Option func(cmd *cobra.Command)

// PrependPersistentPreRun prepends fn to the PersistentPreRun function of the command.
func PrependPersistentPreRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPreRun = prependFunc(cmd.PersistentPreRun, fn...)
	}
}

// PrependPersistentPreRunE prepends fn to the PersistentPreRunE function of the command.
func PrependPersistentPreRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPreRunE = prependFuncE(cmd.PersistentPreRunE, fn...)
	}
}

// PrependPersistentPostRun prepends fn to the PersistentPostRun function of the command.
func PrependPersistentPostRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPostRun = prependFunc(cmd.PersistentPostRun, fn...)
	}
}

// PrependPersistentPostRunE prepends fn to the PersistentPostRunE function of the command.
func PrependPersistentPostRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPostRunE = prependFuncE(cmd.PersistentPostRunE, fn...)
	}
}

// PrependPreRun prepends fn to the PreRun function of the command.
func PrependPreRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PreRun = prependFunc(cmd.PreRun, fn...)
	}
}

// PrependPreRunE prepends fn to the PreRunE function of the command.
func PrependPreRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PreRunE = prependFuncE(cmd.PreRunE, fn...)
	}
}

// PrependPostRun prepends fn to the PostRun function of the command.
func PrependPostRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PostRun = prependFunc(cmd.PostRun, fn...)
	}
}

// PrependPostRunE prepends fn to the PostRunE function of the command.
func PrependPostRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PostRunE = prependFuncE(cmd.PostRunE, fn...)
	}
}

// AppendPersistentPreRun appends fn to the PersistentPreRun function of the command.
func AppendPersistentPreRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPreRun = appendFunc(cmd.PersistentPreRun, fn...)
	}
}

// AppendPersistentPreRunE appends fn to the PersistentPreRunE function of the command.
func AppendPersistentPreRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPreRunE = appendFuncE(cmd.PersistentPreRunE, fn...)
	}
}

// AppendPersistentPostRun appends fn to the PersistentPostRun function of the command.
func AppendPersistentPostRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPostRun = appendFunc(cmd.PersistentPostRun, fn...)
	}
}

// AppendPersistentPostRunE appends fn to the PersistentPostRunE function of the command.
func AppendPersistentPostRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PersistentPostRunE = appendFuncE(cmd.PersistentPostRunE, fn...)
	}
}

// AppendPreRun appends fn to the PreRun function of the command.
func AppendPreRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PreRun = appendFunc(cmd.PreRun, fn...)
	}
}

// AppendPreRunE appends fn to the PreRunE function of the command.
func AppendPreRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PreRunE = appendFuncE(cmd.PreRunE, fn...)
	}
}

// AppendPostRun appends fn to the PostRun function of the command.
func AppendPostRun(fn ...func(cmd *cobra.Command, args []string)) Option {
	return func(cmd *cobra.Command) {
		cmd.PostRun = appendFunc(cmd.PostRun, fn...)
	}
}

// AppendPostRunE appends fn to the PostRunE function of the command.
func AppendPostRunE(fn ...func(cmd *cobra.Command, args []string) error) Option {
	return func(cmd *cobra.Command) {
		cmd.PostRunE = appendFuncE(cmd.PostRunE, fn...)
	}
}

func prependFunc(current func(cmd *cobra.Command, args []string), fn ...func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		for i := range fn {
			fn[i](cmd, args)
		}
		if current != nil {
			current(cmd, args)
		}
	}
}

func prependFuncE(current func(cmd *cobra.Command, args []string) error, fn ...func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		for i := range fn {
			if err = fn[i](cmd, args); err != nil {
				return err
			}
		}
		if current != nil {
			return current(cmd, args)
		}
		return nil
	}
}

func appendFunc(current func(cmd *cobra.Command, args []string), fn ...func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if current != nil {
			current(cmd, args)
		}
		for i := range fn {
			fn[i](cmd, args)
		}
	}
}

func appendFuncE(current func(cmd *cobra.Command, args []string) error, fn ...func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		if current != nil {
			if err = current(cmd, args); err != nil {
				return err
			}
		}
		for i := range fn {
			if err = fn[i](cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
