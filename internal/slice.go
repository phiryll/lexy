package internal

import (
	"io"
)

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// Use MakeSliceCodec(Codec[E]) to create a new sliceCodec.
// A slice is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if empty, prefixEmpty
// - if non-empty, prefixNonEmpty followed by its encoded elements
//
// Encoded elements are escaped and termninated if elemCodec requires it.
type sliceCodec[S ~[]E, E any] struct {
	elemCodec   Codec[E]
	writePrefix prefixWriter[S]
}

func SliceCodec[S ~[]E, E any](elemCodec Codec[E], nilsFirst bool) Codec[S] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, E]{
		TerminateIfNeeded(elemCodec),
		getPrefixWriter[S](isNilSlice, nil, nilsFirst),
	}
}

func (c sliceCodec[S, E]) Read(r io.Reader) (S, error) {
	if value, done, err := ReadPrefix[S](r, true, nil); done {
		return value, err
	}
	values := S{}
	for {
		value, err := c.elemCodec.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (c sliceCodec[S, E]) Write(w io.Writer, value S) error {
	if done, err := c.writePrefix(w, value); done {
		return err
	}
	for _, elem := range value {
		if err := c.elemCodec.Write(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func (c sliceCodec[P, E]) RequiresTerminator() bool {
	return true
}
