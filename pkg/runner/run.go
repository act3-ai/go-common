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

	cobra.OnInitialize(func() {
		level.Set(slog.Level(verbosity))
	})

	cmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", config.EnvIntOr(verbosityEnvName, 0),
		"Logging verbosity level (also setable with environment variable "+verbosityEnvName+")")
	x := cmd.PersistentFlags().Lookup("verbosity")
	x.NoOptDefVal = "0"

	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}

// Run will run the root level cobra command but first setup logging with Zap
func Run(cmd *cobra.Command, verbosityEnvName string) error {
	return RunWithContext(context.Background(), cmd, verbosityEnvName)
}
