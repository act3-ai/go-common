// Package test provides test helper functions
package test

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

// Logger constructs a logger that uses tb.Log().
// The verbosity can be changes with the environment variable TEST_VERBOSITY.
func Logger(tb testing.TB, verbosity int) *slog.Logger {
	tb.Helper()
	if levelStr, exists := os.LookupEnv("TEST_VERBOSITY"); exists {
		v, err := strconv.Atoi(levelStr)
		require.NoError(tb, err)
		verbosity = v
	}

	h := slogt.Factory(func(w io.Writer) slog.Handler {
		opts := &slog.HandlerOptions{
			Level: slog.Level(verbosity),
		}
		return slog.NewTextHandler(w, opts)
	})

	return slogt.New(tb, h)
}
