package internal

import (
	"bytes"
	"io"
)

// mapCodec is the Codec for maps, without sorting the keys.
// Use NewMapCodec[K,V](codec[K], codec[V]) to create a new mapCodec.
// A map is encoded as:
//
//   - if nil, nothing
//   - if empty, PrefixEmpty
//   - if non-empty, PrefixNonEmpty followed by its entries separated by (unescaped) delimiters,
//     each entry encoded as [escaped key, delimiter, escaped value]
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
	empty := make(map[K]V)
	if m, done, err := readPrefix[map[K]V](r, true, &empty); done {
		return m, err
	}
	m := make(map[K]V)
	for {
		b, readErr := Unescape(r)
		if readErr != nil && readErr != io.EOF {
			return m, readErr
		}
		key, codecErr := c.keyCodec.Read(bytes.NewBuffer(b))
		if codecErr != nil && codecErr != io.EOF {
			return m, codecErr
		}

		b, readErr = Unescape(r)
		if readErr != nil && readErr != io.EOF {
			return m, readErr
		}
		value, codecErr := c.valueCodec.Read(bytes.NewBuffer(b))
		if codecErr != nil && codecErr != io.EOF {
			return m, codecErr
		}
		m[key] = value
		if readErr == io.EOF {
			break
		}
	}
	if len(m) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return m, nil
}

func isNilMap[K comparable, V any](value map[K]V) bool {
	return value == nil
}

func isEmptyMap[K comparable, V any](value map[K]V) bool {
	// okay to be true for a nil map, nil is tested first
	return len(value) == 0
}

func (c mapCodec[K, V]) Write(w io.Writer, value map[K]V) error {
	if done, err := writePrefix(w, isNilMap, isEmptyMap, value); done {
		return err
	}
	var buf bytes.Buffer
	notFirst := false
	for k, v := range value {
		if notFirst {
			if _, err := w.Write(del); err != nil {
				return err
			}
		}
		notFirst = true

		if err := c.keyCodec.Write(&buf, k); err != nil {
			return err
		}
		if _, err := Escape(w, buf.Bytes()); err != nil {
			return err
		}
		if _, err := w.Write(del); err != nil {
			return err
		}
		buf.Reset()

		if err := c.valueCodec.Write(&buf, v); err != nil {
			return err
		}
		if _, err := Escape(w, buf.Bytes()); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
