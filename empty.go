package lexy

// emptyCodec is a Codec that encodes no data.
// Get returns the zero value of T.
// No method of this Codec will ever fail.
//
// This is useful for empty structs, which are often used as map values.
type emptyCodec[T any] struct{}

func (emptyCodec[T]) Append(buf []byte, _ T) []byte {
	return buf
}

func (emptyCodec[T]) Put(_ []byte, _ T) int {
	return 0
}

func (emptyCodec[T]) Get(_ []byte) (T, int) {
	var zero T
	return zero, 0
}

func (emptyCodec[T]) RequiresTerminator() bool {
	return true
}
