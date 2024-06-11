package internal

import "io"

type MapCodec[K comparable, V any] struct{}

func (c MapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	panic("unimplemented")
}

func (c MapCodec[K, V]) Write(w io.Writer, value map[K]V) error {
	panic("unimplemented")
}
