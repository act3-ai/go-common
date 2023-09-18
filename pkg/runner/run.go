// Package runner provides common bootstrapping functionality for CLI tools
package runner

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// RunWithContext will run the root level cobra command but first setup logging with Zap
func RunWithContext(ctx context.Context, cmd *cobra.Command, verbosityEnvName string) error {

	level := new(slog.LevelVar)
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, options))

	// flags
	var verbosity int

	// Set verbosity in the "OnInitialize" function,
	//  verbosity flag must be parsed before it can be used
	cobra.OnInitialize(func() {
		/*
			slog.Level values from https://pkg.go.dev/log/slog#Level
			const (
				LevelDebug Level = -4
				LevelInfo  Level = 0
				LevelWarn  Level = 4
				LevelError Level = 8
			)

			LevelError is set as the default logging level.

			For slog, "lower" levels mean a chattier logger, so a user is intending to decrease the value of the slog logger's Level when they increase the value of the verbosity flag. Since slog's levels are on multiples of 4, the value of the verbosity flag is multiplied by 4 to easily increase the verbosity to the next level defined. Without the multiplication, a user rerunning a command with a verbosity of 1 to see more logs would see no difference in output, and there is no reason for them to learn the conventions of the Go log/slog package to confidently use our tools.
		*/
		level.Set(slog.LevelError - slog.Level(verbosity*4))
	})

	cmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", config.EnvIntOr(verbosityEnvName, 0),
		"Logging verbosity level (also setable with environment variable "+verbosityEnvName+")\n"+
			"Levels: 0=ERROR, 1=WARN, 2=INFO, 3=DEBUG")
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "1"

	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}

// Run will run the root level cobra command but first setup logging with Zap
func Run(cmd *cobra.Command, verbosityEnvName string) error {
	return RunWithContext(context.Background(), cmd, verbosityEnvName)
}
