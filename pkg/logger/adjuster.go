package logger

import (
	"context"
	"log/slog"
)

// levelAdjustedHandler is an extension of the slog.Handler interface that maintains a log level bias term.  The bias
// value is positive for a quieter (less chatty) log.  The levelAdjustedHandler usually wraps a standard slog handler
// and is typically applied using the static V() function on an existing logger, with a bias term.  The intention is to
// enable the chattiness of a logger to be reduced when entering deeper call levels, and thus requiring greater
// verbosity settings to see the upgraded log.
type levelAdjustedHandler struct {
	slog.Handler
	bias int
}

// Enabled returns true if the requested log level adjusted by the current bias is greater than or equal to
// the handler's current log level.  A positive bias decreases the chattiness, so a higher verbosity level is required
// for the log entry to be output.
func (h *levelAdjustedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level-slog.Level(h.bias))
}

// Handle adds the biased log level to the underlying handler's level.  Since the bias value is positive for increased
// verbosity, this subtracts the bias value from the record level in actuality.
func (h *levelAdjustedHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Level -= slog.Level(h.bias)
	return h.Handler.Handle(ctx, record) //nolint:wrapcheck
}

// TODO maybe NewLevelAdjustedHandler() should not be exported (and then it can return *levelAdjustedHandler).

// NewLevelAdjustedHandler wraps the handler to perform logging at the adjusted
// level.  The bias value is positive decrease chattiness, e.g., by making Info log entries silent unless verbosity has
// been upgraded.
func NewLevelAdjustedHandler(handler slog.Handler, bias int) slog.Handler {
	// if the handler is already a levelAdjustedHander then we replace it
	// instead of always adding a new adapter/handler to the stack as a performance optimization
	newBias := bias
	lah, ok := handler.(*levelAdjustedHandler)
	if ok {
		// optimization
		newBias += lah.bias
		handler = lah.Handler
	}
	return &levelAdjustedHandler{
		Handler: handler,
		bias:    newBias,
	}
}

// V alters the log level of the provided logger, returning a new logger that will perform logging at the adjusted
// level.  The bias value is positive decrease chattiness, e.g., by making Info log entries silent unless verbosity has
// been upgraded.
func V(log *slog.Logger, bias int) *slog.Logger {
	h := NewLevelAdjustedHandler(log.Handler(), bias)
	return slog.New(h)
}
