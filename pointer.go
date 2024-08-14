package lexy

import (
	"io"
)

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its referent.
// A pointer is encoded as:
//   - if nil, prefixNilFirst/Last
//   - if non-nil, prefixNonNil followed by its encoded referent
type pointerCodec[E any] struct {
	elemCodec Codec[E]
	nilsFirst bool
}

func (c pointerCodec[E]) Read(r io.Reader) (*E, error) {
	if done, err := ReadPrefix(r); done {
		return nil, err
	}
	value, err := c.elemCodec.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	return &value, nil
}

func (c pointerCodec[E]) Write(w io.Writer, value *E) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	return c.elemCodec.Write(w, *value)
}

func (c pointerCodec[E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}

func (c pointerCodec[E]) NilsLast() NillableCodec[*E] {
	return pointerCodec[E]{c.elemCodec, false}
}
