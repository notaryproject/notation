package ioutil

import "io"

// CountWriter counts the written bytes
type CountWriter struct {
	W io.Writer
	N int64
}

// NewCountWriter generates a new writer
func NewCountWriter(w io.Writer) *CountWriter {
	return &CountWriter{W: w}
}

func (w *CountWriter) Write(p []byte) (n int, err error) {
	n, err = w.W.Write(p)
	w.N += int64(n)
	return
}
