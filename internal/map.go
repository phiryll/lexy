package internal

import (
	"io"
)

// mapCodec is the unordered Codec for maps.
// Use MakeMapCodec(Codec[K], Codec[V], bool) to create a new mapCodec.
// A map is encoded as:
//
// - if nil, PrefixNil
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty, encoded key, encoded value, encoded key, ...
//
// Encoded keys and values are escaped and termninated if their respective Codecs require it.
type mapCodec[M ~map[K]V, K comparable, V any] struct {
	keyCodec   Codec[K]
	valueCodec Codec[V]
}

func MapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return mapCodec[M, K, V]{keyCodec, valueCodec}
}

func isNilMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value == nil
}

func isEmptyMap[M ~map[K]V, K comparable, V any](value M) bool {
	// okay to be true for a nil map, nil is tested first
	return len(value) == 0
}

func (c mapCodec[M, K, V]) Read(r io.Reader) (M, error) {
	empty := make(M)
	if m, done, err := readPrefix(r, true, &empty); done {
		return m, err
	}
	keyReader := TerminateIfNeeded(c.keyCodec)
	valueReader := TerminateIfNeeded(c.valueCodec)
	m := make(M)
	for {
		key, err := keyReader.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, err
		}
		value, err := valueReader.Read(r)
		if err != nil {
			return m, err
		}
		m[key] = value
	}
	if len(m) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return m, nil
}

func (c mapCodec[M, K, V]) Write(w io.Writer, value M) error {
	if done, err := writePrefix(w, isNilMap, isEmptyMap, value); done {
		return err
	}
	keyWriter := TerminateIfNeeded(c.keyCodec)
	valueWriter := TerminateIfNeeded(c.valueCodec)
	for k, v := range value {
		if err := keyWriter.Write(w, k); err != nil {
			return err
		}
		if err := valueWriter.Write(w, v); err != nil {
			return err
		}
	}
	return nil
}

func (c mapCodec[M, K, V]) RequiresTerminator() bool {
	return true
}
