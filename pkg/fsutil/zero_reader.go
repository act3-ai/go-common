package fsutil

import (
	"fmt"
	"io"
)

// NewZeroReader returns an io.Reader that always returns zeros but stops with EOF after size bytes
func NewZeroReader(size int64) (io.Reader, error) {
	if size < 0 {
		return nil, fmt.Errorf("size cannot be negative")
	}
	return io.LimitReader(zeroReader{}, size), nil
}

// zeroReader is a reader that always returns zeros.
type zeroReader struct{}

// Read that always returns zeros.
func (z zeroReader) Read(p []byte) (int, error) {
	const zeroByte = 0
	for i := range p {
		p[i] = zeroByte
	}
	return len(p), nil
}
