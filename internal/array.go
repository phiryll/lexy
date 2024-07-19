package internal

import (
	"fmt"
	"io"
	"reflect"
)

// arrayCodec is the Codec for arrays, using elemCodec to encode and decode its elements.
// Use MakeArrayCodec[A any](Codec[E]) to create a new arrayCodec (A is the array type).
// Arrays of different sizes are different types in go, and will require different codecs.
// An array is encoded as its encoded elements.
// Encoded elements are escaped and termninated if elemCodec requires it.
//
// arrayCodec makes heavy use of reflection, and should be avoided if possible.
type arrayCodec[A any, E any] struct {
	arrayType reflect.Type
	elemCodec Codec[E]
}

func MakeArrayCodec[A any, E any](elemCodec Codec[E]) Codec[A] {
	arrayType := reflect.TypeFor[A]()
	elemType := reflect.TypeFor[E]()
	if arrayType.Kind() != reflect.Array {
		panic(fmt.Sprintf("not an array type: %s", arrayType.String()))
	}
	if elemType != arrayType.Elem() {
		panic(fmt.Sprintf("expected element type %s, got %s", arrayType.Elem().String(), elemType.String()))
	}
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return arrayCodec[A, E]{arrayType, elemCodec}
}

func (c arrayCodec[A, E]) Read(r io.Reader) (A, error) {
	var zero A
	values := reflect.New(c.arrayType).Elem()
	codec := TerminateIfNeeded(c.elemCodec)
	size := c.arrayType.Len()
	for i := range size {
		value, err := codec.Read(r)
		if err == io.EOF {
			if i != size-1 {
				return zero, io.ErrUnexpectedEOF
			}
			break
		}
		if err != nil {
			return zero, err
		}
		values.Index(i).Set(reflect.ValueOf(value))
	}
	return values.Interface().(A), nil
}

func (c arrayCodec[A, E]) Write(w io.Writer, value A) error {
	codec := TerminateIfNeeded(c.elemCodec)
	arrayValue := reflect.ValueOf(value)
	for i := range c.arrayType.Len() {
		elem := arrayValue.Index(i).Interface()
		if err := codec.Write(w, elem.(E)); err != nil {
			return err
		}
	}
	return nil
}

func (c arrayCodec[A, E]) RequiresTerminator() bool {
	return true
}
