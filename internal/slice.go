package internal

import "io"

type SliceCodec[T any] struct{}

func (c SliceCodec[T]) Read(r io.Reader) ([]T, error) {
	panic("unimplemented")
}

func (c SliceCodec[T]) Write(w io.Writer, value []T) error {
	panic("unimplemented")
}
