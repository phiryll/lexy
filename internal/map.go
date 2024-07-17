package internal

import (
	"bytes"
	"io"
	"slices"
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

// Similar to mapCodec, except for a Codec ordered by the map's key encodings.
type orderedMapCodec[M ~map[K]V, K comparable, V any] struct {
	keyWriter  Writer[K]
	pairReader pairReader[K, V]
	pairWriter pairWriter[[]byte, V]
}

func MakeMapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return mapCodec[M, K, V]{
		pairReader[K, V]{keyCodec, valueCodec},
		pairWriter[K, V]{keyCodec, valueCodec},
	}
}

func MakeOrderedMapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return orderedMapCodec[M, K, V]{
		keyCodec,
		pairReader[K, V]{keyCodec, valueCodec},
		pairWriter[[]byte, V]{bytesWriter, valueCodec},
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

func (c orderedMapCodec[M, K, V]) Read(r io.Reader) (M, error) {
	return readMap[M](r, c.pairReader)
}

func (c orderedMapCodec[M, K, V]) Write(w io.Writer, value M) error {
	if done, err := writePrefix(w, isNilMap, isEmptyMap, value); done {
		return err
	}

	// It would be cleaner to sort a slice of encoded key-value pairs,
	// but this will be more space-efficient if value encodings are large.
	// OTOH, if keys are large, this may be worse since we're creating a copy of each key.
	type keyBytes struct {
		key K
		b   []byte
	}
	sorted := make([]keyBytes, len(value))
	i := 0
	for key := range value {
		// We can't reuse this buffer, buf.Bytes() is shared.
		var buf bytes.Buffer
		if err := c.keyWriter.Write(&buf, key); err != nil {
			return err
		}
		sorted[i] = keyBytes{key, buf.Bytes()}
		i++
	}
	slices.SortFunc(sorted, func(a, b keyBytes) int {
		return bytes.Compare(a.b, b.b)
	})

	// The rest is very similar to mapCodec.Write
	var scratch bytes.Buffer
	for _, kb := range sorted {
		if err := c.pairWriter.write(w, kb.b, value[kb.key], &scratch); err != nil {
			return err
		}
	}
	return nil
}

func (c mapCodec[M, K, V]) RequiresTerminator() bool {
	return true
}

func (c orderedMapCodec[M, K, V]) RequiresTerminator() bool {
	return true
}
