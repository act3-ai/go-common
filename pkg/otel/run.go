package otel

import (
	"context"
	"fmt"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/bridges/otelslog"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
)

// RunWithContext will run the root level cobra command, with an OpenTelemetry
// configuration.
func RunWithContext(ctx context.Context, cmd *cobra.Command, cfg *Config, verbosityEnvName string) error {
	var err error
	ctx, err = Init(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initializing OpenTelemetry: %w", err)
	}

	handlers := []slog.Handler{runner.SetupLoggingHandler(cmd, verbosityEnvName)}
	if cfg.logProvider != nil {
		handlers = append(handlers, otelslog.NewHandler(cmd.Name(), otelslog.WithLoggerProvider(cfg.logProvider)))
	}

	multiHandler := slogmulti.Fanout(handlers...)
	log := slog.New(multiHandler)
	ctx = logger.NewContext(ctx, log)
	return cmd.ExecuteContext(ctx)
}
