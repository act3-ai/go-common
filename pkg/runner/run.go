// Package runner provides common bootstrapping functionality for CLI tools
package runner

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// RunWithContext will run the root level cobra command but first setup logging with Zap
func RunWithContext(ctx context.Context, cmd *cobra.Command, verbosityEnvName string) error {
	handler := SetupLoggingHandler(cmd, verbosityEnvName)
	log := slog.New(handler)
	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}

// Run will run the root level cobra command but first setup logging with Zap
func Run(cmd *cobra.Command, verbosityEnvName string) error {
	return RunWithContext(context.Background(), cmd, verbosityEnvName)
}
