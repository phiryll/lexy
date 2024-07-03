package internal

import "io"

type mapCodec[K comparable, V any] struct {
	keyCodec   codec[K]
	valueCodec codec[V]
}

func NewMapCodec[K comparable, V any](keyCodec codec[K], valueCodec codec[V]) codec[map[K]V] {
	// TODO: use default if possible based on types
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return mapCodec[K, V]{keyCodec, valueCodec}
}

func (c mapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	panic("unimplemented")
}

func (c mapCodec[K, V]) Write(w io.Writer, value map[K]V) error {
	panic("unimplemented")
}
