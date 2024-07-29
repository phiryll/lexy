package internal

import (
	"fmt"
	"io"
)

// Prefixes to use for encodings that would normally encode nil or an empty value as zero bytes.
// The values were chosen so that nils-first < empty < non-empty < nils-last,
// and the prefixes don't need to be escaped.
//
// This is normally only an issue for variable length encodings.
//
// This prevents ambiguous encodings like these
// (0x00 is the terminator after slice elements, if required):
//
//	""                     => []
//
//	[]string(nil)          => []
//	[]string{}             => []
//
//	[][]string{{}, {}}     => [0x00, 0x00]
//	[][]string{{}, {""}}   => [0x00, 0x00]
//	[][]string{{""}, {}}   => [0x00, 0x00]
//	[][]string{{""}, {""}} => [0x00, 0x00]
const (
	prefixNilFirst byte = 0x02
	prefixEmpty    byte = 0x03
	prefixNonEmpty byte = 0x04
	prefixNilLast  byte = 0x05

	ExportForTestingPrefixNilFirst = prefixNilFirst
	ExportForTestingPrefixEmpty    = prefixEmpty
	ExportForTestingPrefixNonEmpty = prefixNonEmpty
	ExportForTestingPrefixNilLast  = prefixNilLast
)

// Convenience byte slices.
var (
	pNilFirst = []byte{prefixNilFirst}
	pEmpty    = []byte{prefixEmpty}
	pNonEmpty = []byte{prefixNonEmpty}
	pNilLast  = []byte{prefixNilLast}
)

func isNilPointer[P ~*E, E any](value P) bool {
	return value == nil
}

func isNilSlice[S ~[]E, E any](value S) bool {
	return value == nil
}

func isNilMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value == nil
}

func isEmptyString[T ~string](value T) bool {
	return len(value) == 0
}

func isEmptySlice[S ~[]E, E any](value S) bool {
	return value != nil && len(value) == 0
}

func isEmptyMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value != nil && len(value) == 0
}

// ReadPrefix reads the prefix byte from r and handles nil and empty values.
// nilable should be true if and only if nil is an allowed value of type T.
// emptyValue should point to the empty value of type T if it differs from the zero value of T.
//
// ReadPrefix returns done == false only if the value itself still needs to be read
// (value is neither nil nor empty), and there was no error reading the prefix.
// If ReadPrefix returns done == true and err is nil, the returned value is either nil or the empty value.
// ReadPrefix will never return an error value of io.EOF.
//
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
func ReadPrefix[T any](r io.Reader, nilable bool, emptyValue *T) (value T, done bool, err error) {
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
	case prefixNilFirst, prefixNilLast:
		if !nilable {
			return zero, true, fmt.Errorf("read nil for non-nilable type %T", zero)
		}
		if err == io.EOF {
			// ignore EOF
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

// The signature of WritePrefixNilsFirst/Last without the isNil and isEmpty arguments.
// Used to simplify code using getPrefixWriter below, see pointer.go for a usage example.
type prefixWriter[T any] func(w io.Writer, value T) (done bool, err error)

func getPrefixWriter[T any](isNil, isEmpty func(T) bool, nilsFirst bool) prefixWriter[T] {
	if nilsFirst {
		return func(w io.Writer, value T) (done bool, err error) {
			return WritePrefixNilsFirst(w, isNil, isEmpty, value)
		}
	}
	return func(w io.Writer, value T) (done bool, err error) {
		return WritePrefixNilsLast(w, isNil, isEmpty, value)
	}
}

// WritePrefixNilsFirst writes the correct prefix byte for value to w, with nils ordered first.
// isNil or isEmpty should be non-nil if type T allows nil or empty values respectively.
//
// WritePrefixNilsFirst returns done == false only if the value itself still needs to be written
// (value is neither nil nor empty), and there was no error writing the prefix.
// If WritePrefixNilsFirst returns done == true and err is nil,
// the value was nil or empty and no further data needs to be written for this value.
//
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
func WritePrefixNilsFirst[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	if isNil != nil && isNil(value) {
		_, err := w.Write(pNilFirst)
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

// Exactly the same as WritePrefixNilsFirst, except nils are ordered last.
func WritePrefixNilsLast[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	if isNil != nil && isNil(value) {
		_, err := w.Write(pNilLast)
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
