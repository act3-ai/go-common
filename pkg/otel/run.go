package otel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/act3-ai/go-common/pkg/config/env"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/runner"
)

// Run will run the root level cobra command, with logging and with the provided
// OpenTelemetry configuration.
func Run(ctx context.Context, cmd *cobra.Command, cfg *Config, verbosityEnvName string) error {
	if env.BoolOr("OTEL_INSTRUMENTATION_ENABLED", false) {
		// Run root command with OTel instrumentation enabled.
		return run(ctx, cmd, cfg, verbosityEnvName)
	}
	return runner.Run(ctx, cmd, verbosityEnvName)
}

func run(ctx context.Context, cmd *cobra.Command, cfg *Config, verbosityEnvName string) error {
	if cfg == nil {
		cfg = &Config{} // ensure to check for environment configuration
	}
	var err error
	ctx, err = cfg.Init(ctx)
	if err != nil {
		return fmt.Errorf("initializing OpenTelemetry providers: %w", err)
	}
	defer func() {
		if err := cfg.Shutdown(ctx); err != nil {
			slog.WarnContext(ctx, "OTEL shutdown failed", "error", err)
		}
	}()

	// create a single log handler with a handler for stderr and otel
	stderrHandler := runner.SetupLoggingHandler(cmd, verbosityEnvName)
	otelWrappedHandler := cfg.WrapSlogHandler(cmd.Name(), stderrHandler)

	log := slog.New(otelWrappedHandler)
	slog.SetDefault(log)
	ctx = logger.NewContext(ctx, log)

	// errors from cfg.Shutdown() are not fatal so we just log them
	return cmd.ExecuteContext(ctx)
}
