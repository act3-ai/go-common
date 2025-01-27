package otel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
)

// RunWithContext will run the root level cobra command, with the provided
// OpenTelemetry configuration. It calls cfg.Init and cfg.Close appropriately.
func RunWithContext(ctx context.Context, cmd *cobra.Command, cfg *Config, verbosityEnvName string) error {
	if cfg == nil {
		cfg = &Config{} // ensure to check for environment configuration
	}
	var err error
	ctx, err = cfg.Init(ctx)
	if err != nil {
		return fmt.Errorf("initializing OpenTelemetry providers: %w", err)
	}
	defer cfg.Shutdown(ctx)

	// create a single log handler with a handler for stderr and otel
	stderrHandler := runner.SetupLoggingHandler(cmd, verbosityEnvName)
	otelWrappedHandler := cfg.WrapSlogHandler(cmd.Name(), stderrHandler)

	log := slog.New(otelWrappedHandler)
	ctx = logger.NewContext(ctx, log)

	return cmd.ExecuteContext(ctx)
}
