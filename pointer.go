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
	prefix    Prefix
}

func (c pointerCodec[E]) Append(buf []byte, value *E) []byte {
	return AppendUsingWrite[*E](c, buf, value)
}

func (c pointerCodec[E]) Put(buf []byte, value *E) int {
	return PutUsingAppend[*E](c, buf, value)
}

func (c pointerCodec[E]) Get(buf []byte) (*E, int) {
	return GetUsingRead[*E](c, buf)
}

func (c pointerCodec[E]) Write(w io.Writer, value *E) error {
	if done, err := c.prefix.Write(w, value == nil); done {
		return err
	}
	return c.elemCodec.Write(w, *value)
}

func (c pointerCodec[E]) Read(r io.Reader) (*E, error) {
	if done, err := c.prefix.Read(r); done {
		return nil, err
	}
	value, err := c.elemCodec.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	return &value, nil
}

func (c pointerCodec[E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}

func (c pointerCodec[E]) NilsLast() NillableCodec[*E] {
	return pointerCodec[E]{c.elemCodec, PrefixNilsLast}
}
