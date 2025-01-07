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

const otelErrKey = "otelErr"

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

	slogRouter := slogmulti.Router()
	slogRouter.Add(runner.SetupLoggingHandler(cmd, verbosityEnvName), allowAll())
	if cfg.logProvider != nil {
		// bridge slog to the log provider
		otelHandler := otelslog.NewHandler(cmd.Name(), otelslog.WithLoggerProvider(cfg.logProvider))
		slogRouter.Add(otelHandler, ignoreOtel())
	}

	log := slog.New(slogRouter.Handler())
	ctx = logger.NewContext(ctx, log)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		// attribute tells slogRounter to not send these logs to telemetry, i.e. logged only locally
		// TODO: Ideally, we would be able to filter out only the failures to emit log telemetry.
		// This filtering is only necessary to prevent the log telemetry emit recursion (see ignoreOtel).
		logger.FromContext(ctx).With(otelErrKey, "otel_emit_err").ErrorContext(ctx, "failed to emit telemetry", "error", err)
	}))

	return cmd.ExecuteContext(ctx)
}

// ignoreOtel filters out error logs provided by the otel ErrorHandler
// indicating a log telemetry emit error, ensuring they're only emitted locally. It
// prevents infinite recursion in case log telemetry fails, e.g.
// emit log telemetry --> failure to send logs to endpoint -->
// log telemetry emit err --> emit log telemetry --> ...
func ignoreOtel() func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
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

func allowAll() func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return true
	}
}
