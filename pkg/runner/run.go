// Package runner provides common bootstrapping functionality for CLI tools
package runner

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Run will run the root level cobra command but first setup logging
func Run(ctx context.Context, cmd *cobra.Command, verbosityEnvName string) error {
	handler := SetupLoggingHandler(cmd, verbosityEnvName)
	log := slog.New(handler)
	slog.SetDefault(log)
	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}
