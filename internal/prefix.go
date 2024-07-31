package internal

import (
	"fmt"
	"io"
)

// Prefixes to use for encodings for types whose instances can be nil.
// The values were chosen so that nils-first < non-nil < nils-last,
// and neither the prefixes nor their complements need to be escaped.
const (
	// Room for more between non-nil and nils-last if needed.
	prefixNilFirst byte = 0x02
	prefixNonNil   byte = 0x03
	prefixNilLast  byte = 0xFD

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

// ReadPrefix reads the nil/non-nil prefix byte from r and returns which it read.
//
// ReadPrefix returns done == false only if the non-nil value still needs to be read,
// and there was no error reading the prefix.
//
// If ReadPrefix returns done == true, then the caller is done reading this value
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was read.
//
// ReadPrefix will return io.EOF only if no bytes were read and r.Read returned io.EOF.
// ReadPrefix will not return an error if a prefix was successfully read and r.Read returned io.EOF,
// because the read of the prefix was successful.
// Any subsequent read from r by the caller will properly return 0 bytes read and io.EOF.
func ReadPrefix(r io.Reader) (done bool, err error) {
	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		// We must propagate io.EOF in this case.
		if err == nil {
			err = fmt.Errorf("unexpected read of 0 bytes with no error")
		}
		return true, err
	}
	// If we successfully read a byte and get io.EOF, ignore the EOF.
	if err == io.EOF {
		err = nil
	} else if err != nil {
		return true, err
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

// WritePrefix writes a nil/non-nil prefix byte to w based on the values of isNil and nilsFirst.
//
// WritePrefix returns done == false only if isNil is false and there was no error writing the prefix,
// in which case the caller still needs to write the non-nil value to w.
//
// If WritePrefix returns done == true, then the caller is done writing the current value to w
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was successfully written.
func WritePrefix(w io.Writer, isNil, nilsFirst bool) (done bool, err error) {
	var prefix []byte
	switch {
	case !isNil:
		prefix = pNonNil
	case nilsFirst:
		prefix = pNilFirst
	default:
		prefix = pNilLast
	}
	if _, err := w.Write(prefix); err != nil {
		return true, err
	}
	return isNil, nil
}
