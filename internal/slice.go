package internal

import (
	"io"
)

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// Use MakeSliceCodec(Codec[E]) to create a new sliceCodec.
// A slice is encoded as:
//
// - if nil, PrefixNil
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty followed by its encoded elements
//
// Encoded elements are escaped and termninated if elemCodec requires it.
type sliceCodec[S ~[]E, E any] struct {
	elemCodec Codec[E]
}

func SliceCodec[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, E]{elemCodec}
}

func (c sliceCodec[S, E]) Read(r io.Reader) (S, error) {
	empty := S{}
	if value, done, err := readPrefix(r, true, &empty); done {
		return value, err
	}
	codec := TerminateIfNeeded(c.elemCodec)
	var values S
	for {
		value, err := codec.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}
	if len(values) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return values, nil
}

func (c sliceCodec[S, E]) Write(w io.Writer, value S) error {
	if done, err := writePrefix(w, isNilSlice, isEmptySlice, value); done {
		return err
	}
	codec := TerminateIfNeeded(c.elemCodec)
	for _, elem := range value {
		if err := codec.Write(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func (c sliceCodec[P, E]) RequiresTerminator() bool {
	return true
}
