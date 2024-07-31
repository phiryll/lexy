package internal

import (
	"io"
	"slices"
)

// negateCodec negates codec, reversing the ordering of its encoding.
//
// Every encoding will be greater than any prefix of that encoding (definition of lexicographical ordering).
// For example, consider these encodings:
//
//	A = {0x00, 0x02, 0x03}
//	B = {0x00, 0x02, 0x03, 0x00}
//	A < B
//
// This Codec must effectively reverse that for what the delegate codec produces.
// Just flipping all the bits works except when one encoding is the prefix of another.
// The above example with all bits flipped is:
//
//	^A = {0xFF, 0xFD, 0xFC}
//	^B = {0xFF, 0xFD, 0xFC, 0xFF}
//
// We need to transform these results so that -B is less than -A.
// Adding a 0xFF terminator accomplishes this,
// but then we have another escape/terminator problem, just with 0xFF and 0xFE instead of 0x00 and 0x01.
// We can achieve the same effect by always escaping and terminating the normal way,
// and then flip all the bits, inluding the trailing terminator.
// If we do that for the above example, we get the correctly negated ordering.
//
//	esc+term(A) = {0x01, 0x00, 0x02, 0x03, 0x00}
//	esc+term(B) = {0x01, 0x00, 0x02, 0x03, 0x01, 0x00, 0x00}
//
//	^esc+term(A) = {0xFE, 0xFF, 0xFD, 0xFC, 0xFF}
//	^esc+term(B) = {0xFE, 0xFF, 0xFD, 0xFC, 0xFE, 0xFF, 0xFF}
type negateCodec[T any] struct {
	codec Codec[T]
}

func NegateCodec[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	// Negate must escape and terminate its delegate whether it requries it or not,
	// but shouldn't wrap if the delegate is already a terminatorCodec.
	if _, ok := codec.(terminatorCodec[T]); !ok {
		codec = Terminate(codec)
	}
	return negateCodec[T]{codec}
}

func (c negateCodec[T]) Read(r io.Reader) (T, error) {
	return c.codec.Read(negateReader{r})
}

func (c negateCodec[T]) Write(w io.Writer, value T) error {
	return c.codec.Write(negateWriter{w}, value)
}

func (c negateCodec[T]) RequiresTerminator() bool {
	return false
}

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
	negate(p[:n])
	return n, err
}

// negateWriter is an io.Writer which flips all the bits,
// negating in the sense in the sense of lexicographical ordering.
// The argument []byte is cloned and the clone's bits are flipped,
// because delegate Writer is assumed to be more efficient writing the entire slice
// than it would be writing multiple smaller slices to avoid the allocation.
type negateWriter struct {
	io.Writer
}

func (w negateWriter) Write(p []byte) (int, error) {
	b := slices.Clone(p)
	negate(b)
	return w.Writer.Write(b)
}
