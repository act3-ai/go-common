package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLevelAdjustedHandler(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		base := slog.NewTextHandler(nil, nil)
		got := NewLevelAdjustedHandler(base, 2)
		require.IsType(t, &levelAdjustedHandler{}, got)
		lah := got.(*levelAdjustedHandler)
		require.Equal(t, base, lah.Handler)
		assert.Equal(t, lah.bias, 2)
	})

	t.Run("nested", func(t *testing.T) {
		base := slog.NewTextHandler(nil, nil)
		handler := NewLevelAdjustedHandler(base, 2)
		got := NewLevelAdjustedHandler(handler, 3)
		require.IsType(t, &levelAdjustedHandler{}, got)
		lah := got.(*levelAdjustedHandler)
		require.Equal(t, base, lah.Handler)
		assert.Equal(t, lah.bias, 5)
	})
}

func TestV(t *testing.T) {
	ctx := context.Background()

	t.Run("simple", func(t *testing.T) {
		buf := &bytes.Buffer{}
		base := slog.New(slog.NewTextHandler(buf, nil))
		base4 := V(base, 4)
		base6 := V(base4, 2)

		ensureLog := func(log *slog.Logger, level slog.Level, msg string) {
			t.Helper()
			assert.True(t, log.Enabled(ctx, level))
			log.Log(ctx, level, msg)
			assert.Contains(t, buf.String(), msg)
		}

		ensureNoLog := func(log *slog.Logger, level slog.Level, msg string) {
			t.Helper()
			assert.False(t, log.Enabled(ctx, level))
			log.Log(ctx, level, msg)
			assert.NotContains(t, buf.String(), msg)
		}

		ensureLog(base, slog.LevelInfo, "base-A")
		ensureNoLog(base, slog.LevelDebug, "base-B")

		ensureLog(base4, slog.LevelWarn, "base4-A")
		ensureNoLog(base4, slog.LevelInfo, "base4-B")
		ensureNoLog(base4, slog.LevelDebug, "base4-C")

		ensureNoLog(base6, 5, "base6-5")
		ensureLog(base6, 6, "base6-6")
		ensureLog(base6, 7, "base6-7")
		ensureLog(base6, 6, "base6-6")
	})
}
