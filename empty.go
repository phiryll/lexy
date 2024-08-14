package lexy

import "io"

// emptyCodec is a Codec that writes and reads no data.
// Read returns the zero value of T.
// Read and Write will never return an error, including io.EOF.
//
// This is useful for empty structs, which are often used as map values.
type emptyCodec[T any] struct{}

func (emptyCodec[T]) Read(_ io.Reader) (T, error) {
	var zero T
	return zero, nil
}

func (emptyCodec[T]) Write(_ io.Writer, _ T) error {
	return nil
}

func (emptyCodec[T]) RequiresTerminator() bool {
	return true
}
