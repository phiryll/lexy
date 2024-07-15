package internal

import (
	"io"
)

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its pointee.
// Use MakePointerCodec(Codec[T]) to create a new pointerCodec.
// A pointer is encoded as:
//
// - if nil, nothing
// - if non-nil, PrefixNonEmpty followed by its encoded pointee
//
// The prefix is required to disambiguate a nil pointer from a pointer to a nil value.
type pointerCodec[T any] struct {
	elemCodec Codec[T]
}

func MakePointerCodec[T any](elemCodec Codec[T]) Codec[*T] {
	// TODO: Might want 2 implementations based on T,
	// whether values require escaping and delimiting or not.
	// Does whether elemCodec requires termination answer that question?

	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return pointerCodec[T]{elemCodec}
}

func (c pointerCodec[T]) Read(r io.Reader) (*T, error) {
	if ptr, done, err := readPrefix[*T](r, true, nil); done {
		return ptr, err
	}
	value, err := c.elemCodec.Read(r)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func isNilPointer[T any](value *T) bool {
	return value == nil
}

func (c pointerCodec[T]) Write(w io.Writer, value *T) error {
	if done, err := writePrefix(w, isNilPointer, nil, value); done {
		return err
	}
	return c.elemCodec.Write(w, *value)
}
