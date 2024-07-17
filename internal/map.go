package internal

import (
	"bytes"
	"io"
)

// mapCodec is the unordered Codec for maps.
// Use MakeMapCodec(Codec[K], Codec[V], bool) to create a new mapCodec.
// A map is encoded as:
//
//   - if nil, nothing
//   - if empty, PrefixEmpty
//   - if non-empty,
//     [ PrefixNonEmpty,
//     escaped encoded key, delimiter, escaped encoded value, delimiter,
//     escaped encoded key, delimiter, escaped encoded value, delimiter,
//     ...
//     escaped encoded key, delimiter, escaped encoded value]
type mapCodec[M ~map[K]V, K comparable, V any] struct {
	pairReader pairReader[K, V]
	pairWriter pairWriter[K, V]
}

func MakeMapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return mapCodec[M, K, V]{
		pairReader[K, V]{keyCodec, valueCodec},
		pairWriter[K, V]{keyCodec, valueCodec},
	}
}

func isNilMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value == nil
}

func isEmptyMap[M ~map[K]V, K comparable, V any](value M) bool {
	// okay to be true for a nil map, nil is tested first
	return len(value) == 0
}

// Read implementation is the same for unordered and ordered encodings.
func readMap[M ~map[K]V, K comparable, V any](r io.Reader, pairReader pairReader[K, V]) (M, error) {
	empty := make(M)
	if m, done, err := readPrefix(r, true, &empty); done {
		return m, err
	}
	m := make(M)
	for {
		key, value, err := pairReader.read(r)
		if err == io.EOF {
			break
		}
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

func (c mapCodec[M, K, V]) Read(r io.Reader) (M, error) {
	return readMap[M](r, c.pairReader)
}

func (c mapCodec[M, K, V]) Write(w io.Writer, value M) error {
	if done, err := writePrefix(w, isNilMap, isEmptyMap, value); done {
		return err
	}
	var scratch bytes.Buffer
	for k, v := range value {
		if err := c.pairWriter.write(w, k, v, &scratch); err != nil {
			return err
		}
	}
	return nil
}

func (c mapCodec[M, K, V]) RequiresTerminator() bool {
	return true
}
