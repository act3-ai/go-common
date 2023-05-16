package fsutil

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewZeroReader(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected []byte
		wantErr  bool
	}{
		{
			name:     "Zero size",
			size:     0,
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "Positive size",
			size:     10,
			expected: bytes.Repeat([]byte{0}, 10),
			wantErr:  false,
		},
		{
			name:     "Negative size",
			size:     -1,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewZeroReader(tt.size)
			assert.Equal(t, tt.wantErr, err != nil, "unexpected error")

			if err == nil {
				var buf bytes.Buffer
				_, err := io.Copy(&buf, r)
				assert.NoError(t, err, "unexpected error")

				assert.Equal(t, tt.expected, buf.Bytes(), "unexpected output")
			}
		})
	}
}
