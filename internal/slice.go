package internal

import (
	"io"
)

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// Use MakeSliceCodec(Codec[T]) to create a new sliceCodec.
// A slice is encoded as:
//
// - if nil, nothing
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty followed by its encoded elements
// Each element is escaped and termninated if elemCodec requires termination.
type sliceCodec[S ~[]T, T any] struct {
	elemCodec Codec[T]
}

func MakeSliceCodec[S ~[]T, T any](elemCodec Codec[T]) Codec[S] {
	// TODO: Might want 2 implementations based on T,
	// whether elements require escaping and delimiting or not.
	// Does whether elemCodec requires termination answer that question?
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, T]{elemCodec}
}

func (c sliceCodec[S, T]) Read(r io.Reader) (S, error) {
	empty := S{}
	if value, done, err := readPrefix(r, true, &empty); done {
		return value, err
	}
	codec := TerminateIfNeeded(c.elemCodec)
	var values S
	for {
		value, codecErr := codec.Read(r)
		if codecErr == io.EOF {
			break
		}
		if codecErr != nil {
			return values, codecErr
		}
		values = append(values, value)
	}
	if len(values) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return values, nil
}

func isNilSlice[S ~[]T, T any](value S) bool {
	return value == nil
}

func isEmptySlice[S ~[]T, T any](value S) bool {
	// okay to be true for a nil slice, nil is tested first
	return len(value) == 0
}

func (c sliceCodec[S, T]) Write(w io.Writer, value S) error {
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

func (c sliceCodec[P, T]) RequiresTerminator() bool {
	return true
}
