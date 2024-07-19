package internal

import (
	"bytes"
	"io"
	"slices"
)

// negateCodec negates codec, reversing the ordering of its encoding.
// Use MakeNegateCodec(Codec[T]) to create a new negateCodec.
//
// negateCodec must escape and terminate when encoding,
// because otherwise it wouldn't know when to stop reading when decoding.
type negateCodec[T any] struct {
	// This implementation is essentially the same as terminator,
	// but with bit flipping, and being thread-safe if codec is.
	codec Codec[T]
}

func MakeNegateCodec[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	return negateCodec[T]{codec}
}

func (c negateCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	b, readErr := unescape(r)
	if readErr != nil && (readErr != io.EOF || len(b) == 0) {
		return value, readErr
	}
	negate(b)
	value, codecErr := c.codec.Read(bytes.NewBuffer(b))
	if codecErr != nil && codecErr != io.EOF {
		return value, codecErr
	}
	return value, nil
}

func (c negateCodec[T]) Write(w io.Writer, value T) error {
	var scratch bytes.Buffer
	if err := c.codec.Write(&scratch, value); err != nil {
		return err
	}
	b := scratch.Bytes()
	negate(b)
	if _, err := escape(w, b); err != nil {
		return err
	}
	return nil
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
