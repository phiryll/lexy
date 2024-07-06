package internal

import "io"

type structCodec[T, V any] struct {
	fieldCodec Codec[V]
}

func NewStructCodec[T, V any](fieldCodec Codec[V]) Codec[T] {
	// TODO: use default if possible based on V
	if fieldCodec == nil {
		panic("fieldCodec must be non-nil")
	}
	return structCodec[T, V]{fieldCodec}
}

func (c structCodec[T, F]) Read(r io.Reader) (T, error) {
	panic("unimplemented")
}

func (com structCodec[T, F]) Write(w io.Writer, value T) error {
	panic("unimplemented")
}
