package lexy

import (
	"errors"
	"io"
)

// mapCodec is the unordered Codec for maps.
// A map is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil, encoded key, encoded value, encoded key, ...
//
// Encoded keys and values are escaped and termninated if their respective Codecs require it.
type mapCodec[K comparable, V any] struct {
	keyCodec   Codec[K]
	valueCodec Codec[V]
	nilsFirst  bool
}

func (c mapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	if done, err := ReadPrefix(r); done {
		return nil, err
	}
	m := map[K]V{}
	for {
		key, err := c.keyCodec.Read(r)
		if errors.Is(err, io.EOF) {
			return m, nil
		}
		if err != nil {
			return m, err
		}
		value, err := c.valueCodec.Read(r)
		if err != nil {
			return m, UnexpectedIfEOF(err)
		}
		m[key] = value
	}
}

func (c mapCodec[K, V]) Write(w io.Writer, value map[K]V) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	for k, v := range value {
		if err := c.keyCodec.Write(w, k); err != nil {
			return err
		}
		if err := c.valueCodec.Write(w, v); err != nil {
			return err
		}
	}
	return nil
}

func (mapCodec[K, V]) RequiresTerminator() bool {
	return true
}

func (c mapCodec[K, V]) NilsLast() NillableCodec[map[K]V] {
	return mapCodec[K, V]{c.keyCodec, c.valueCodec, false}
}
