package test

import (
	"os"
	"strconv"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
)

// GetTestLogger constructs a test logger. The verbosity can be changes with the environment variable TEST_VERBOSITY
func GetTestLogger(t *testing.T, verbosity int) logr.Logger {
	t.Helper()
	if levelStr, exists := os.LookupEnv("TEST_VERBOSITY"); exists {
		v, err := strconv.Atoi(levelStr)
		require.NoError(t, err)
		verbosity = v
	}
	return testr.NewWithOptions(t, testr.Options{
		Verbosity: verbosity,
	})
}
