package ioutil

// WriterCounter (type int64) is an implementation of io.Writer.
// The counter is incremented by the length of each write call.
// Recommended to be used with io.MultiWriter.
type WriterCounter int64

// Write implements the io.Writer interface.
// It increments the counter by the length of the input.
func (c *WriterCounter) Write(p []byte) (int, error) {
	n := len(p)
	*c += WriterCounter(n)
	return n, nil
}
