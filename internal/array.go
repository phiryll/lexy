package internal

import (
	"fmt"
	"io"
	"reflect"
)

// pointerToArrayCodec is the Codec for pointers to arrays, using elemCodec to encode and decode its elements.
// Use MakePointerToArrayCodec[P ~*A, A any](Codec[E]) to create a new arrayCodec (A is the array type).
// Arrays of different sizes are different types in go, and will require different codecs.
// An array is encoded as its encoded elements.
// Encoded elements are escaped and termninated if elemCodec requires it.
//
// pointerToArrayCodec makes heavy use of reflection, and should be avoided if possible.
type pointerToArrayCodec[P ~*A, A any, E any] struct {
	pointerType reflect.Type
	arrayType   reflect.Type
	elemCodec   Codec[E]
	writePrefix prefixWriter[P]
}

// arrayCodec is the Codec for arrays, using elemCodec to encode and decode its elements.
// Use MakeArrayCodec[A any](Codec[E]) to create a new arrayCodec (A is the array type).
// Arrays of different sizes are different types in go, and will require different codecs.
// An array is encoded as its encoded elements.
// Encoded elements are escaped and termninated if elemCodec requires it.
//
// arrayCodec makes heavy use of reflection, and should be avoided if possible.
//
// This codec delegates to a pointerToArrayCodec internally.
type arrayCodec[A any, E any] struct {
	delegate Codec[*A]
}

func PointerToArrayCodec[P ~*A, A any, E any](elemCodec Codec[E], nilsFirst bool) Codec[P] {
	pointerType := reflect.TypeFor[P]()
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
	return pointerToArrayCodec[P, A, E]{
		pointerType,
		arrayType,
		elemCodec,
		getPrefixWriter[P](isNilPointer, nil, nilsFirst),
	}
}

func ArrayCodec[A any, E any](elemCodec Codec[E]) Codec[A] {
	return arrayCodec[A, E]{PointerToArrayCodec[*A, A, E](elemCodec, true)}
}

func (c pointerToArrayCodec[P, A, E]) Read(r io.Reader) (P, error) {
	if ptr, done, err := ReadPrefix[P](r, true, nil); done {
		return ptr, err
	}
	ptrToPtrToArray := reflect.New(c.pointerType)
	ptrToPtrToArray.Elem().Set(reflect.New(c.arrayType))
	array := ptrToPtrToArray.Elem().Elem()
	codec := TerminateIfNeeded(c.elemCodec)
	size := c.arrayType.Len()
	for i := range size {
		value, err := codec.Read(r)
		if err == io.EOF {
			if i != size-1 {
				return nil, io.ErrUnexpectedEOF
			}
			break
		}
		if err != nil {
			return nil, err
		}
		array.Index(i).Set(reflect.ValueOf(value))
	}
	return ptrToPtrToArray.Elem().Interface().(P), nil
}

func (c pointerToArrayCodec[P, A, E]) Write(w io.Writer, value P) error {
	if done, err := c.writePrefix(w, value); done {
		return err
	}
	codec := TerminateIfNeeded(c.elemCodec)
	arrayValue := reflect.ValueOf(value).Elem()
	for i := range c.arrayType.Len() {
		elem := arrayValue.Index(i).Interface()
		if err := codec.Write(w, elem.(E)); err != nil {
			return err
		}
	}
	return nil
}

func (c pointerToArrayCodec[P, A, E]) RequiresTerminator() bool {
	return false
}

func (c arrayCodec[A, E]) Read(r io.Reader) (A, error) {
	var zero A
	ptrToValue, err := c.delegate.Read(r)
	if err != nil {
		return zero, err
	}
	return *ptrToValue, nil
}

func (c arrayCodec[A, E]) Write(w io.Writer, value A) error {
	return c.delegate.Write(w, &value)
}

func (c arrayCodec[A, E]) RequiresTerminator() bool {
	return false
}
