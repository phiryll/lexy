package internal

import "io"

type MapCodec[K comparable, V any] struct{}

func (c MapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	panic("unimplemented")
}

func (c MapCodec[K, V]) Write(value map[K]V, w io.Writer) error {
	panic("unimplemented")
}
