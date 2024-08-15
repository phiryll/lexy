package lexy

import (
	"errors"
	"fmt"
)

var (
	errUnexpectedNilsFirst = errors.New("read nils-first prefix when nils-last was configured")
	errUnexpectedNilsLast  = errors.New("read nils-last prefix when nils-first was configured")
	errBigFloatEncoding    = errors.New("unexpected failure encoding big.Float")
)

type unknownPrefixError struct {
	prefix byte
}

func (e unknownPrefixError) Error() string {
	return fmt.Sprintf("unexpected prefix %X", e.prefix)
}

type bufferTooSmallError struct {
	bufSize, requiredSize int
}

func (e bufferTooSmallError) Error() string {
	return fmt.Sprintf("[]byte with length %d cannot hold %d bytes", e.bufSize, e.requiredSize)
}

// checkBufferSize panics with an errBufferTooSmall if buf can't hold n bytes.
func checkBufferSize(buf []byte, n int) {
	if len(buf) < n {
		panic(bufferTooSmallError{len(buf), n})
	}
}

type nilError struct {
	name string
}

func (e nilError) Error() string {
	return e.name + " must be non-nil"
}

// checkNonNil panics if x is nil.
// This should only be used if there isn't a simpler way to raise a panic,
// like accessing a field or invoking a method.
func checkNonNil(x any, name string) {
	if x == nil {
		panic(nilError{name})
	}
}
