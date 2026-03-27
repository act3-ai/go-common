package basicenc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicEncoding(t *testing.T) {
	// enc stores the encoding rules for JSON Pointers.
	enc := NewBasicEncoding([][2]string{
		{"~", "~0"},
		{"/", "~1"},
	})

	tests := []struct {
		name    string
		raw     string
		escaped string
	}{
		{
			name:    "no special characters",
			raw:     "hello",
			escaped: "hello",
		},
		{
			name:    "empty string",
			raw:     "",
			escaped: "",
		},
		{
			name:    "tilde only",
			raw:     "~",
			escaped: "~0",
		},
		{
			name:    "slash only",
			raw:     "/",
			escaped: "~1",
		},
		{
			name:    "both tilde and slash",
			raw:     "~/",
			escaped: "~0~1",
		},
		{
			name:    "slash then tilde",
			raw:     "/~",
			escaped: "~1~0",
		},
		{
			name:    "multiple tildes",
			raw:     "~~",
			escaped: "~0~0",
		},
		{
			name:    "multiple slashes",
			raw:     "//",
			escaped: "~1~1",
		},
		{
			name:    "complex string with path-like structure",
			raw:     "a/b~c",
			escaped: "a~1b~0c",
		},
		{
			name:    "string with mixed characters",
			raw:     "hello/world~test",
			escaped: "hello~1world~0test",
		},
		{
			name:    "RFC6901 example - a/b",
			raw:     "a/b",
			escaped: "a~1b",
		},
		{
			name:    "RFC6901 example - m~n",
			raw:     "m~n",
			escaped: "m~0n",
		},
		{
			name:    "edge case - ~01 sequence",
			raw:     "~01",
			escaped: "~001",
		},
		{
			name:    "edge case - ~10 sequence",
			raw:     "~10",
			escaped: "~010",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Encode", func(t *testing.T) {
				got := enc.Encode(tt.raw)
				assert.Equal(t, tt.escaped, got)
			})
			t.Run("Decode", func(t *testing.T) {
				got := enc.Decode(tt.escaped)
				assert.Equal(t, tt.raw, got)
			})
		})
	}
}
