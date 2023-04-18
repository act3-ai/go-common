package fsutil

// zeroReader is a reader that always returns zeros.
type zeroReader struct{}

// Read that always returns zeros.
func (z zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
