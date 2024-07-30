package internal

import (
	"io"
)

// mapCodec is the unordered Codec for maps.
// A map is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil, encoded key, encoded value, encoded key, ...
//
// Encoded keys and values are escaped and termninated if their respective Codecs require it.
type mapCodec[M ~map[K]V, K comparable, V any] struct {
	keyCodec    Codec[K]
	valueCodec  Codec[V]
	writePrefix prefixWriter[M]
}

func MapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V], nilsFirst bool) Codec[M] {
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return mapCodec[M, K, V]{
		TerminateIfNeeded(keyCodec),
		TerminateIfNeeded(valueCodec),
		getPrefixWriter[M](isNilMap, nilsFirst),
	}
}

func (c mapCodec[M, K, V]) Read(r io.Reader) (M, error) {
	if isNil, err := ReadPrefix(r); isNil {
		return nil, err
	}
	m := make(M)
	for {
		key, err := c.keyCodec.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, err
		}
		value, err := c.valueCodec.Read(r)
		if err != nil {
			return m, unexpectedIfEOF(err)
		}
		m[key] = value
	}
	return m, nil
}

func (c mapCodec[M, K, V]) Write(w io.Writer, value M) error {
	if done, err := c.writePrefix(w, value); done {
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

func (c mapCodec[M, K, V]) RequiresTerminator() bool {
	return true
}
