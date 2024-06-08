package internal

import "io"

type StructCodec[T any] struct{}

func (c StructCodec[T]) Read(r io.Reader) (T, error) {
	panic("unimplemented")
}

func (com StructCodec[T]) Write(value T, w io.Writer) error {
	panic("unimplemented")
}
