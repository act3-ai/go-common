package otel

import (
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

// WrapSlogHandler produces a slog.Handler that writes logs to OpenTelemetry and the base slog.Handler.
// Base handler is optional.
func (c *Config) WrapSlogHandler(name string, base slog.Handler) slog.Handler {
	if c.logProvider == nil {
		// Return unmodified base handler.
		return base
	}

	// bridge slog to the log provider, which adds traceid and spanid's to the log
	otelHandler := otelslog.NewHandler(name, otelslog.WithLoggerProvider(c.logProvider))

	if base == nil {
		// Return otelslog handler if no base handler provided.
		return otelHandler
	}

	// create a single logger with a handler for base and otel
	slogRouter := slogmulti.Router().Add(base).Add(otelHandler)

	// Any telemetry error is simply logged as it shouldn't be fatal.
	// To avoid having multiple loggers in the context, we "fork" the logs to
	// multiple handlers via slogRouter. As a result,  we end up not having
	// access to a logger as early as we want. Thus, we wait to set the error
	// handler and shutdown until after the logger is created; which required
	// the telemetry logger provider to already be initialized.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		// log otel errors to base handler directly, skipping the router so they are only logged locally
		// without this, errors could produce an infinite recursion of errors
		slog.New(base).Error("failed to emit telemetry", slog.Any("error", err))
	}))

	return slogRouter.Handler()
}
