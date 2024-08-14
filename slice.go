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
	nilsFirst bool
}

func (c sliceCodec[E]) Read(r io.Reader) ([]E, error) {
	if done, err := ReadPrefix(r); done {
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

func (c sliceCodec[E]) Write(w io.Writer, value []E) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	for _, elem := range value {
		if err := c.elemCodec.Write(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func (sliceCodec[E]) RequiresTerminator() bool {
	return true
}

func (c sliceCodec[E]) NilsLast() NillableCodec[[]E] {
	return sliceCodec[E]{c.elemCodec, false}
}
