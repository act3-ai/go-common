package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
)

// Run will run the root level cobra command, with logging and with the provided
// OpenTelemetry configuration.
func Run(ctx context.Context, cmd *cobra.Command, cfg *Config, verbosityEnvName string) error {
	if cfg == nil {
		cfg = &Config{} // ensure to check for environment configuration
	}
	var err error
	ctx, err = cfg.Init(ctx)
	if err != nil {
		return fmt.Errorf("initializing OpenTelemetry providers: %w", err)
	}
	defer cfg.Shutdown(ctx) //nolint:errcheck

	// create a single log handler with a handler for stderr and otel
	stderrHandler := runner.SetupLoggingHandler(cmd, verbosityEnvName)
	otelWrappedHandler := cfg.WrapSlogHandler(cmd.Name(), stderrHandler)

	log := slog.New(otelWrappedHandler)
	ctx = logger.NewContext(ctx, log)

	return errors.Join(cmd.ExecuteContext(ctx), cfg.Shutdown(ctx)) //nolint:wrapcheck
}
