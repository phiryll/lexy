package lexy

import (
	"errors"
	"io"
)

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// A slice is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil followed by its encoded elements
//
// Encoded elements are escaped and termninated if elemCodec requires it.
type sliceCodec[E any] struct {
	elemCodec Codec[E]
	prefix    Prefix
}

func (c sliceCodec[E]) Append(buf []byte, value []E) []byte {
	return AppendUsingWrite[[]E](c, buf, value)
}

func (c sliceCodec[E]) Put(buf []byte, value []E) int {
	return PutUsingAppend[[]E](c, buf, value)
}

func (c sliceCodec[E]) Get(buf []byte) ([]E, int) {
	return GetUsingRead[[]E](c, buf)
}

func (c sliceCodec[E]) Write(w io.Writer, value []E) error {
	if done, err := c.prefix.Write(w, value == nil); done {
		return err
	}
	for _, elem := range value {
		if err := c.elemCodec.Write(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func (c sliceCodec[E]) Read(r io.Reader) ([]E, error) {
	if done, err := c.prefix.Read(r); done {
		return nil, err
	}
	values := []E{}
	for {
		value, err := c.elemCodec.Read(r)
		if errors.Is(err, io.EOF) {
			return values, nil
		}
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}
}

func (sliceCodec[E]) RequiresTerminator() bool {
	return true
}

func (c sliceCodec[E]) NilsLast() NillableCodec[[]E] {
	return sliceCodec[E]{c.elemCodec, PrefixNilsLast}
}
