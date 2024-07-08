package internal

import (
	"fmt"
	"io"
)

// The function signatures of Codec.Read and Codec.Write.
// These are useful for implementing some Codecs.
type (
	Reader[T any] func(io.Reader) (T, error)
	Writer[T any] func(io.Writer, T) error
)

// Same interface as lexy.Codec, to avoid a circular dependency.
// lexy.Codec cannot be a type alias to this, because generic type aliases are not permitted.
type Codec[T any] interface {
	Read(r io.Reader) (T, error)
	Write(w io.Writer, value T) error
}

// implementation of writer[[]byte] that just writes the bytes
func writeBytes(w io.Writer, b []byte) error {
	_, err := w.Write(b)
	return err
}

var _ Writer[[]byte] = writeBytes

func unexpectedIfEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Prefixes, documented in lexy.go
const (
	// 0x02 is reserved for nil if that becomes necessary.
	PrefixEmpty    byte = 0x03
	PrefixNonEmpty byte = 0x04
)

// Convenience byte slices.
var (
	prefixEmpty    = []byte{PrefixEmpty}
	prefixNonEmpty = []byte{PrefixNonEmpty}
)

// Reads the prefix and handles nil and empty values.
// nilable should be true if and only if nil is an allowed value of type T.
// emptyValue should point to the empty value of type T if it differs from the zero value of T.
// Returns done = false only if the value itself still needs to be read
// and there was no error reading the prefix.
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
func readPrefix[T any](r io.Reader, nilable bool, emptyValue *T) (value T, done bool, err error) {
	// nil for types that can be nil (slices, maps, pointers)
	// empty value for non-nil types that can be empty (string)
	// non-nil, non-empty zero value otherwise (bool, int8, ...)
	var zero T

	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		// 0 bytes read
		if !nilable && (err == nil || err == io.EOF) {
			// cannot be nil, 0 bytes read is always an error
			err = io.ErrUnexpectedEOF
		} else if err == io.EOF {
			// no EOF if nil is allowed
			err = nil
		}
		return zero, true, err
	}
	switch prefix[0] {
	case PrefixEmpty:
		if err != nil && err != io.EOF {
			return zero, true, err
		}
		// ignore EOF
		if emptyValue != nil {
			return *emptyValue, true, nil
		}
		return zero, true, nil
	case PrefixNonEmpty:
		if err == io.EOF {
			return zero, true, io.ErrUnexpectedEOF
		}
		return zero, false, err
	default:
		if err == nil || err == io.EOF {
			err = fmt.Errorf("unexpected prefix %X", prefix[0])
		}
		return zero, true, err
	}
}

// Writes the correct prefix for value, or nothing if the value is nil.
// isNil or isEmpty should be non-nil if type T allows nil or empty values respectively.
// isEmpty is used after isNil, so isEmpty can also return true for nil values.
// Returns done = false only if the value itself still needs to be written
// and there was no error writing the prefix.
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
func writePrefix[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	if isNil != nil && isNil(value) {
		// do nothing
		return true, nil
	}
	if isEmpty != nil && isEmpty(value) {
		_, err := w.Write(prefixEmpty)
		return true, err
	}
	if _, err := w.Write(prefixNonEmpty); err != nil {
		return true, err
	}
	return false, nil
}
