package internal

import "io"

type StructCodec[T any] struct{}

func (c StructCodec[T]) Read(r io.Reader) (T, error) {
	panic("unimplemented")
}

func (com StructCodec[T]) Write(w io.Writer, value T) error {
	panic("unimplemented")
}
