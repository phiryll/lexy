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
	prefix     Prefix
}

func (c mapCodec[K, V]) Append(buf []byte, value map[K]V) []byte {
	return AppendUsingWrite[map[K]V](c, buf, value)
}

func (c mapCodec[K, V]) Put(buf []byte, value map[K]V) int {
	return PutUsingAppend[map[K]V](c, buf, value)
}

func (c mapCodec[K, V]) Get(buf []byte) (map[K]V, int) {
	return GetUsingRead[map[K]V](c, buf)
}

func (c mapCodec[K, V]) Write(w io.Writer, value map[K]V) error {
	if done, err := c.prefix.Write(w, value == nil); done {
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

func (c mapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	if done, err := c.prefix.Read(r); done {
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

func (mapCodec[K, V]) RequiresTerminator() bool {
	return true
}

func (c mapCodec[K, V]) NilsLast() NillableCodec[map[K]V] {
	return mapCodec[K, V]{c.keyCodec, c.valueCodec, PrefixNilsLast}
}
