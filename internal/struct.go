package internal

import "io"

type structCodec[T, V any] struct {
	fieldCodec Codec[V]
}

func MakeStructCodec[T, V any](fieldCodec Codec[V]) Codec[T] {
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

func (c structCodec[T, F]) RequiresTerminator() bool {
	// should only be true if some field requires it,
	// but we can't figure that out here.
	return true
}
