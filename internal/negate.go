package internal

import (
	"io"
	"slices"
)

// Negate negates b, in the sense of lexicographical ordering.
func negate(b []byte) {
	for i := range b {
		b[i] ^= 0xFF
	}
}

var (
	_ io.Reader = negateReader{}
	_ io.Writer = negateWriter{}
)

// negateReader is an io.Reader which flips all the bits,
// negating in the sense in the sense of lexicographical ordering.
type negateReader struct {
	io.Reader
}

func (r negateReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	negate(p)
	return n, err
}

// negateWriter is an io.Writer which flips all the bits,
// negating in the sense in the sense of lexicographical ordering.
type negateWriter struct {
	io.Writer
}

func (w negateWriter) Write(p []byte) (int, error) {
	b := slices.Clone(p)
	negate(b)
	return w.Writer.Write(b)
}
