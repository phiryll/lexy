package internal

import (
	"bytes"
	"io"
)

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// Use MakeSliceCodec(Codec[T]) to create a new sliceCodec.
// A slice is encoded as:
//
// - if nil, nothing
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty followed by its elements encoded and escaped separated by (unescaped) delimiters
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

func readElements[S ~[]T, T any](r io.Reader, elemCodec Codec[T]) (S, error) {
	var values S
	for {
		value, codecErr := elemCodec.Read(r)
		if codecErr == io.EOF {
			return values, nil
		}
		if codecErr != nil {
			return values, codecErr
		}
		values = append(values, value)
	}
}

func readDelimitedElements[S ~[]T, T any](r io.Reader, elemCodec Codec[T]) (S, error) {
	var values S
	for {
		b, readErr := Unescape(r)
		if readErr != nil && readErr != io.EOF {
			return values, readErr
		}
		value, codecErr := elemCodec.Read(bytes.NewBuffer(b))
		if codecErr != nil && codecErr != io.EOF {
			return values, codecErr
		}
		values = append(values, value)
		if readErr == io.EOF {
			break
		}
	}
	return values, nil
}

func (c sliceCodec[S, T]) Read(r io.Reader) (S, error) {
	empty := S{}
	if value, done, err := readPrefix(r, true, &empty); done {
		return value, err
	}
	var values S
	var err error
	if c.elemCodec.RequiresTerminator() {
		values, err = readDelimitedElements[S](r, c.elemCodec)
	} else {
		values, err = readElements[S](r, c.elemCodec)
	}
	if err != nil {
		return values, err
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

func writeElements[S ~[]T, T any](w io.Writer, elemCodec Codec[T], value S) error {
	for _, elem := range value {
		if err := elemCodec.Write(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func writeDelimitedElements[S ~[]T, T any](w io.Writer, elemCodec Codec[T], value S) error {
	var scratch bytes.Buffer
	for i, elem := range value {
		if i > 0 {
			if _, err := w.Write(del); err != nil {
				return err
			}
		}
		scratch.Reset()
		if err := elemCodec.Write(&scratch, elem); err != nil {
			return err
		}
		if _, err := Escape(w, scratch.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (c sliceCodec[S, T]) Write(w io.Writer, value S) error {
	if done, err := writePrefix(w, isNilSlice, isEmptySlice, value); done {
		return err
	}
	if c.elemCodec.RequiresTerminator() {
		return writeDelimitedElements(w, c.elemCodec, value)
	} else {
		return writeElements(w, c.elemCodec, value)
	}
}

func (c sliceCodec[P, T]) RequiresTerminator() bool {
	return true
}
