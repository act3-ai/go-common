package otel

import (
	"context"
	"fmt"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"

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
	defer cfg.Shutdown()

	// create a single logger with a handler for stderr and otel
	slogRouter := slogmulti.Router()
	slogRouter = slogRouter.Add(runner.SetupLoggingHandler(cmd, verbosityEnvName), allowAll())
	if cfg.logProvider != nil {
		// bridge slog to the log provider
		otelHandler := otelslog.NewHandler(cmd.Name(), otelslog.WithLoggerProvider(cfg.logProvider))
		slogRouter = slogRouter.Add(otelHandler, ignoreOtel())
	}

	log := slog.New(slogRouter.Handler())
	ctx = logger.NewContext(ctx, log)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		// attribute tells slogRounter to not send these logs to telemetry, i.e. logged only locally
		logger.FromContext(ctx).With(otelErrKey, "otel_emit_err").ErrorContext(ctx, "failed to emit telemetry", "error", err)
	}))

	return cmd.ExecuteContext(ctx)
}

const otelErrKey = "otelErr"

// ignoreOtel filters out error logs provided by the otel ErrorHandler
// indicating a log telemetry emit error, ensuring they're only emitted locally.
// It prevents infinite recursion in case log telemetry fails, e.g.
// emit log telemetry --> failure to send logs to endpoint -->
// log telemetry emit err --> emit log telemetry --> ...
func ignoreOtel() func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		// TODO: Ideally, we filter out only failures to emit log telemetry. but
		// this is not trivial.
		ok := true
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == otelErrKey {
				ok = false
				return false
			}
			return true
		})
		return ok
	}
}

// allowAll performs no log filtering.
func allowAll() func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return true
	}
}
