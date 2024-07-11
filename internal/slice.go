package internal

import (
	"bytes"
	"io"
)

// sliceCodec is the Codec for slices, using elementCodec to encode and decode its elements.
// Use NewSliceCodec(Codec[T]) to create a new sliceCodec.
// A slice is encoded as:
//
// - if nil, nothing
// - if empty, PrefixEmpty
// - if non-empty, PrefixNonEmpty followed by its elements encoded and escaped separated by (unescaped) delimiters
type sliceCodec[T any] struct {
	elementCodec Codec[T]
}

func NewSliceCodec[T any](elementCodec Codec[T]) Codec[[]T] {
	// TODO: use default if possible based on T
	//
	// TODO: Might want 2 implementations based on T,
	// whether elements require escaping and delimiting or not.
	// Does whether elementCodec requires termination answer that question?

	if elementCodec == nil {
		panic("elementCodec must be non-nil")
	}
	return sliceCodec[T]{elementCodec}
}

func (c sliceCodec[T]) Read(r io.Reader) ([]T, error) {
	empty := []T{}
	if value, done, err := readPrefix(r, true, &empty); done {
		return value, err
	}
	var values []T
	for {
		b, readErr := Unescape(r)
		if readErr != nil && readErr != io.EOF {
			return values, readErr
		}
		value, codecErr := c.elementCodec.Read(bytes.NewBuffer(b))
		if codecErr != nil && codecErr != io.EOF {
			return values, codecErr
		}
		values = append(values, value)
		if readErr == io.EOF {
			break
		}
	}
	if len(values) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return values, nil
}

func isNilSlice[T any](value []T) bool {
	return value == nil
}

func isEmptySlice[T any](value []T) bool {
	// okay to be true for a nil slice, nil is tested first
	return len(value) == 0
}

func (c sliceCodec[T]) Write(w io.Writer, value []T) error {
	if done, err := writePrefix(w, isNilSlice, isEmptySlice, value); done {
		return err
	}
	var scratch bytes.Buffer
	for i, value := range value {
		if i > 0 {
			if _, err := w.Write(del); err != nil {
				return err
			}
		}
		scratch.Reset()
		if err := c.elementCodec.Write(&scratch, value); err != nil {
			return err
		}
		if _, err := Escape(w, scratch.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
