// Package logger is common logging functionality to use slog
package logger

import (
	"context"
	"log/slog"
)

type contextKey struct{}

var loggerContextKey contextKey

// FromContext returns the logger stored in the context. If a logger is not found, the default logger will be returned
func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		slog.Debug("did not find logger in context, returning default logger")
		return slog.Default()
	}

	return logger
}

// NewContext wraps the given logger in the given context so that it may be retrieved with FromContext
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}
