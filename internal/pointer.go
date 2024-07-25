package internal

import (
	"io"
)

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its pointee.
// Use MakePointerCodec[P](Codec[E]) for P with underlying type *E to create a new pointerCodec.
// A pointer is encoded as:
//
// - if nil, PrefixNil
// - if non-nil, PrefixNonEmpty followed by its encoded pointee
//
// The prefix is required to disambiguate a nil pointer from a pointer to a nil value.
type pointerCodec[P ~*E, E any] struct {
	elemCodec Codec[E]
}

func PointerCodec[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return pointerCodec[P, E]{elemCodec}
}

func (c pointerCodec[P, E]) Read(r io.Reader) (P, error) {
	if ptr, done, err := ReadPrefix[P](r, true, nil); done {
		return ptr, err
	}
	value, err := c.elemCodec.Read(r)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (c pointerCodec[P, E]) Write(w io.Writer, value P) error {
	if done, err := WritePrefixNilsFirst(w, isNilPointer, nil, value); done {
		return err
	}
	return c.elemCodec.Write(w, *value)
}

func (c pointerCodec[P, E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}
