package timefmt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_unmarshalTimeJSON(t *testing.T) {
	t.Run("JSON null", func(t *testing.T) {
		got := time.Time{}
		err := timeUnmarshalJSON(string(RFC3339UTCDate), nullBytes, &got)
		assert.NoError(t, err)
		assert.True(t, got.IsZero(), "IsZero()")
	})
	t.Run("Not a JSON string", func(t *testing.T) {
		got := time.Time{}
		err := timeUnmarshalJSON(string(RFC3339UTCDate), []byte(`{an object}`), &got)
		assert.Error(t, err)
		assert.True(t, got.IsZero(), "IsZero()")
	})
	t.Run("Fail to parse timestamp", func(t *testing.T) {
		got := time.Time{}
		err := timeUnmarshalJSON(string(RFC3339UTCDate), []byte(`"invalid timestamp"`), &got)
		assert.Error(t, err)
		assert.True(t, got.IsZero(), "IsZero()")
	})
	t.Run("Success", func(t *testing.T) {
		got := time.Time{}
		err := timeUnmarshalJSON(string(RFC3339UTCDate), []byte(`"2026-01-01"`), &got)
		assert.NoError(t, err)
		assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), got)
	})
}
