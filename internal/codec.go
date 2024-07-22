package internal

import (
	"bytes"
	"fmt"
	"io"
)

// Same interface as lexy.Reader, to avoid a circular dependency.
// lexy.Reader cannot be a type alias to this, because generic type aliases are not permitted.
type Reader[T any] interface {
	Read(io.Reader) (T, error)
}

// Same interface as lexy.Writer, to avoid a circular dependency.
// lexy.Writer cannot be a type alias to this, because generic type aliases are not permitted.
type Writer[T any] interface {
	Write(io.Writer, T) error
}

// Same interface as lexy.Codec, to avoid a circular dependency.
// lexy.Codec cannot be a type alias to this, because generic type aliases are not permitted.
type Codec[T any] interface {
	Reader[T]
	Writer[T]
	RequiresTerminator() bool
}

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write when encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(&b, value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read when decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	// bytes.NewBuffer takes ownership of its argument, so we need to clone it first.
	return codec.Read(bytes.NewBuffer(bytes.Clone(data)))
}

func isNilPointer[T any](value *T) bool {
	return value == nil
}

func unexpectedIfEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Prefixes to use for encodings that would normally encode nil or an empty value as zero bytes.
// The values were chosen so that nil < empty < non-empty, and the prefixes don't need to be escaped.
// This is normally only an issue for variable length encodings.
//
// This prevents ambiguous encodings like these
// (0x00 is the terminator between slice elements, if required):
//
//	""                     => []
//
//	[]string{}             => []
//	[]string{""}           => []
//
//	[][]string{{}, {}}     => [0x00]
//	[][]string{{}, {""}}   => [0x00]
//	[][]string{{""}, {}}   => [0x00]
//	[][]string{{""}, {""}} => [0x00]
const (
	prefixNil      byte = 0x02
	prefixEmpty    byte = 0x03
	prefixNonEmpty byte = 0x04

	ExportForTestingPrefixNil      = prefixNil
	ExportForTestingPrefixEmpty    = prefixEmpty
	ExportForTestingPrefixNonEmpty = prefixNonEmpty
)

// Convenience byte slices.
var (
	pNil      = []byte{prefixNil}
	pEmpty    = []byte{prefixEmpty}
	pNonEmpty = []byte{prefixNonEmpty}
)

// Reads the prefix and handles nil and empty values.
// nilable should be true if and only if nil is an allowed value of type T.
// emptyValue should point to the empty value of type T if it differs from the zero value of T.
// Returns done = false only if the value itself still needs to be read (neither nil nor empty),
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
		return zero, true, io.ErrUnexpectedEOF
	}
	switch prefix[0] {
	case prefixNil:
		if !nilable {
			return zero, true, fmt.Errorf("read nil for non-nilable type %T", zero)
		}
		if err == io.EOF {
			// no EOF if nil is allowed
			err = nil
		}
		return zero, true, err
	case prefixEmpty:
		if err != nil && err != io.EOF {
			return zero, true, err
		}
		// ignore EOF
		if emptyValue != nil {
			return *emptyValue, true, nil
		}
		return zero, true, nil
	case prefixNonEmpty:
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

// Writes the correct prefix for value.
// isNil or isEmpty should be non-nil if type T allows nil or empty values respectively.
// isEmpty is used after isNil, so isEmpty can also return true for nil values.
// Returns done = false only if the value itself still needs to be written (neither nil nor empty),
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
		_, err := w.Write(pNil)
		return true, err
	}
	if isEmpty != nil && isEmpty(value) {
		_, err := w.Write(pEmpty)
		return true, err
	}
	if _, err := w.Write(pNonEmpty); err != nil {
		return true, err
	}
	return false, nil
}
