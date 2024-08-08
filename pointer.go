package lexy

import (
	"io"
)

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its pointee.
// A pointer is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil followed by its encoded pointee
//
// The prefix is required to disambiguate a nil pointer from a pointer to a nil value.
type pointerCodec[P ~*E, E any] struct {
	elemCodec Codec[E]
	nilsFirst bool
}

func (c pointerCodec[P, E]) Read(r io.Reader) (P, error) {
	if done, err := ReadPrefix(r); done {
		return nil, err
	}
	value, err := c.elemCodec.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	return &value, nil
}

func (c pointerCodec[P, E]) Write(w io.Writer, value P) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	return c.elemCodec.Write(w, *value)
}

func (c pointerCodec[P, E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}

func (c pointerCodec[P, E]) NilsLast() NillableCodec[P] {
	return pointerCodec[P, E]{c.elemCodec, false}
}
