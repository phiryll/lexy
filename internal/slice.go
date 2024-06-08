package internal

import "io"

type SliceCodec[T any] struct{}

func (c SliceCodec[T]) Read(r io.Reader) ([]T, error) {
	panic("unimplemented")
}

func (c SliceCodec[T]) Write(value []T, w io.Writer) error {
	panic("unimplemented")
}
