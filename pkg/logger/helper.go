package logger

import (
	"context"
	"log/slog"
)

// A log level to always log.  A sufficiently high level.
const levelAlwaysLog = 100

// Error is a helper to log an error.  The record is always logged.
func Error(log *slog.Logger, err error, msg string, args ...any) {
	log.Log(context.Background(), 100, msg, args...) //nolint:sloglint
}

// ErrorContext is a helper to log an error.  The record is always logged.
func ErrorContext(ctx context.Context, log *slog.Logger, err error, msg string, args ...any) {
	log.Log(ctx, levelAlwaysLog, msg, args...) //nolint:sloglint
}
