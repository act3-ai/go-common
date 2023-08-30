// Package test provides test helper functions
package test

import (
	"log/slog"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetTestLogger constructs a test logger. The verbosity can be changes with the environment variable TEST_VERBOSITY
func GetTestLogger(t *testing.T, verbosity int) *slog.Logger {
	t.Helper()
	if levelStr, exists := os.LookupEnv("TEST_VERBOSITY"); exists {
		v, err := strconv.Atoi(levelStr)
		require.NoError(t, err)
		verbosity = v
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(verbosity)}))
}
