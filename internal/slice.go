package internal

import (
	"bytes"
	"fmt"
	"io"
)

// SliceCodec is the Codec for slices, using elementCodec to encode and decode its elements.
// Use NewSliceCodec[T](codec[T]) to create a new SliceCodec.
// A slice is encoded as:
//
// - if nil, nothing
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty followed by its elements encoded and escaped separated by (unescaped) delimiters
type SliceCodec[T any] struct {
	elementCodec codec[T]
}

func NewSliceCodec[T any](elementCodec codec[T]) SliceCodec[T] {
	// TODO: check nil, use default if possible based on T
	//
	// TODO: Might want 2 implementations based on T,
	// whether elements require escaping and delimiting or not.
	// Does whether elementCodec requires termination answer that question?

	if elementCodec == nil {
		panic("elementCodec must be non-nil")
	}
	return SliceCodec[T]{elementCodec}
}

func (c SliceCodec[T]) Read(r io.Reader) ([]T, error) {
	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		if err == nil || err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	switch prefix[0] {
	case PrefixEmpty:
		if err != nil && err != io.EOF {
			return nil, err
		}
		return []T{}, nil
	case PrefixNonEmpty:
		var values []T
		for {
			// err1 = io.Reader failed, may be EOF
			// err2 = Codec translation failed
			b, err1 := Unescape(r)
			value, err2 := c.elementCodec.Read(bytes.NewBuffer(b))
			values = append(values, value)
			if err1 == io.EOF {
				break
			}
			if err1 != nil {
				return values, err1
			}
			if err2 != nil {
				return values, err2
			}
		}
		return values, nil
	default:
		if err == nil || err == io.EOF {
			err = fmt.Errorf("unexpected prefix %X", prefix[0])
		}
		return nil, err
	}
}

func (c SliceCodec[T]) Write(w io.Writer, values []T) error {
	// Enclosing Codec (if any) will write a trailing delimiter if needed.
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		_, err := w.Write(prefixEmpty)
		return err
	}
	if _, err := w.Write(prefixNonEmpty); err != nil {
		return err
	}
	var buf bytes.Buffer
	for i, value := range values {
		if i != 0 {
			if _, err := w.Write(del); err != nil {
				return err
			}
		}
		if err := c.elementCodec.Write(&buf, value); err != nil {
			return err
		}
		if _, err := Escape(w, buf.Bytes()); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
