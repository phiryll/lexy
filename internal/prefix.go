package internal

import (
	"fmt"
	"io"
)

// Prefixes to use for encodings that would normally encode nil value as zero bytes.
// The values were chosen so that nils-first < non-nil < nils-last,
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
	prefixNonNil   byte = 0x04
	prefixNilLast  byte = 0x05

	ExportForTestingPrefixNilFirst = prefixNilFirst
	ExportForTestingPrefixNonNil   = prefixNonNil
	ExportForTestingPrefixNilLast  = prefixNilLast
)

// Convenience byte slices.
var (
	pNilFirst = []byte{prefixNilFirst}
	pNonNil   = []byte{prefixNonNil}
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

// ReadPrefix reads the nil/non-nil prefix byte from r and returns which it read.
//
// If ReadPrefix returns isNil == true, then the caller is done reading this value
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was read.
// ReadPrefix returns isNil == false only if the non-nil value still needs to be read,
// and there was no error reading the prefix.
//
// ReadPrefix will return an error value of io.ErrUnexpectedEOF if no bytes were read.
// ReadPrefix will never return an error value of io.EOF.
func ReadPrefix(r io.Reader) (isNil bool, err error) {
	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		return true, io.ErrUnexpectedEOF
	}
	if err != nil {
		if err != io.EOF {
			return true, err
		}
		// ignore io.EOF
		err = nil
	}
	switch prefix[0] {
	case prefixNilFirst, prefixNilLast:
		return true, nil
	case prefixNonNil:
		return false, nil
	default:
		return true, fmt.Errorf("unexpected prefix %X", prefix[0])
	}
}

// The signature of WritePrefixNilsFirst/Last without the isNil argument.
// Used to simplify code using getPrefixWriter below, see pointer.go for a usage example.
type prefixWriter[T any] func(w io.Writer, value T) (done bool, err error)

func getPrefixWriter[T any](isNil func(T) bool, nilsFirst bool) prefixWriter[T] {
	if nilsFirst {
		return func(w io.Writer, value T) (done bool, err error) {
			return WritePrefixNilsFirst(w, isNil, value)
		}
	}
	return func(w io.Writer, value T) (done bool, err error) {
		return WritePrefixNilsLast(w, isNil, value)
	}
}

// WritePrefixNilsFirst writes the correct prefix byte for value to w, with nils ordered first.
//
// WritePrefixNilsFirst returns done == false only if the value itself still needs to be written
// (value is not nil), and there was no error writing the prefix.
// If WritePrefixNilsFirst returns done == true and err is nil,
// the value was nil and no further data needs to be written for this value.
func WritePrefixNilsFirst[T any](w io.Writer, isNil func(T) bool, value T) (done bool, err error) {
	if isNil(value) {
		_, err := w.Write(pNilFirst)
		return true, err
	}
	if _, err := w.Write(pNonNil); err != nil {
		return true, err
	}
	return false, nil
}

// Exactly the same as WritePrefixNilsFirst, except nils are ordered last.
func WritePrefixNilsLast[T any](w io.Writer, isNil func(T) bool, value T) (done bool, err error) {
	if isNil(value) {
		_, err := w.Write(pNilLast)
		return true, err
	}
	if _, err := w.Write(pNonNil); err != nil {
		return true, err
	}
	return false, nil
}
