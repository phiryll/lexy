package lexy

import (
	"io"
)

// Prefixes to use for encodings for types whose instances can be nil.
// The values were chosen so that prefixNilFirst < prefixNonNil < prefixNilLast,
// and neither the prefixes nor their complements need to be escaped.
const (
	// Room for more between prefixNonNil and prefixNilLast if needed.
	prefixNilFirst byte = 0x02
	prefixNonNil   byte = 0x03
	prefixNilLast  byte = 0xFD
)

// A Prefix provides helper methods to handle the initial nil/non-nil prefix byte
// for [Codec] implementations that encode types whose instances can be nil.
// The rest of these comments only pertain to usage by these [Codec] implementations.
//
// Each Prefix method is a helper for implementing the correspondingly named Codec method.
// Invoking the Prefix method should be the first action taken by the Codec method,
// since it allows an early return if the value is nil.
// The one exception is Get, which should check for an empty argument buffer first.
// All methods process exactly one byte if they are successful.
//
// In addition to other return values, every Prefix method returns done,
// a bool value which is true if and only if the caller should return immediately.
// If done is true, either there was an error or the value is nil.
// If done is false, there was no error and the value is non-nil,
// in which case the caller still needs to process the non-nil value.
// See the method docs for typical usages.
//
// Prefix is implemented by [PrefixNilsFirst] and [PrefixNilsLast].
type Prefix interface {
	// Append appends a prefix byte to the end of buf, returning the updated buffer.
	// This is a typical usage:
	//
	//	func (c fooCodec) Append(buf []byte, value Foo) []byte {
	//	    done, newBuf := c.prefix.Append(buf, value == nil)
	//	    if done {
	//	        return newBuf
	//	    }
	//	    // encode and append the non-nil value to newBuf
	//	}
	Append(buf []byte, isNil bool) (done bool, newBuf []byte)

	// Put sets buf[0] to a prefix byte.
	// This is a typical usage:
	//
	//	func (c fooCodec) Put(buf []byte, value Foo) int {
	//	    if c.prefix.Put(buf, value == nil) {
	//	        return 1
	//	    }
	//	    // encode the non-nil value into buf[1:]
	//	}
	Put(buf []byte, isNil bool) (done bool)

	// Get decodes a prefix byte from buf[0].
	// Get will not modify buf.
	// This is a typical usage:
	//
	//	func (c fooCodec) Get(buf []byte) (Foo, int)
	//	    if len(buf) == 0 {
	//	        return nil, -1
	//	    }
	//	    if c.prefix.Get(buf) {
	//	        return nil, 1
	//	    }
	//	    // decode and return a non-nil value from buf[1:]
	//	}
	Get(buf []byte) (done bool)

	// Write writes a prefix byte to w.
	// This is a typical usage:
	//
	//	func (c fooCodec) Write(w io.Writer, value Foo) error {
	//	    if done, err := c.prefix.Write(w, value == nil); done {
	//	        return err
	//	    }
	//	    // encode and write the non-nil value to w
	//	}
	Write(w io.Writer, isNil bool) (done bool, err error)

	// Read reads a prefix byte from r.
	// This is a typical usage:
	//
	//	func (c fooCodec) Read(r io.Reader) (Foo, error) {
	//	    if done, err := c.prefix.Read(r); done {
	//	        return nil, err
	//	    }
	//	    // read, decode, and return a non-nil value from r
	//	}
	//
	// Read will return [io.EOF] only if no bytes were read and [io.Reader.Read] returned io.EOF.
	// Read will not return an error if a prefix was successfully read and io.Reader.Read returned io.EOF,
	// because the read of the prefix was successful.
	// Any subsequent read from r should properly return 0 bytes read and io.EOF in this case.
	Read(r io.Reader) (done bool, err error)
}

var (
	// PrefixNilsFirst is the [Prefix] implementation ordering nils first.
	PrefixNilsFirst prefixNilsFirst

	// PrefixNilsLast is the [Prefix] implementation ordering nils last.
	PrefixNilsLast prefixNilsLast
)

type (
	prefixNilsFirst struct{}
	prefixNilsLast  struct{}
)

// prefixFor returns which prefix byte to write.
// This method is used by Append, Put, and Write.
func prefixFor(isNil, nilsFirst bool) byte {
	switch {
	case !isNil:
		return prefixNonNil
	case nilsFirst:
		return prefixNilFirst
	default:
		return prefixNilLast
	}
}

// eval returns (done, err), as documented on Prefix, after having read prefix.
// This method is used by Get and Read.
func eval(prefix byte, nilsFirst bool) (bool, error) {
	switch prefix {
	case prefixNonNil:
		return false, nil
	case prefixNilFirst:
		if !nilsFirst {
			return true, errUnexpectedNilsFirst
		}
		return true, nil
	case prefixNilLast:
		if nilsFirst {
			return true, errUnexpectedNilsLast
		}
		return true, nil
	default:
		return true, unknownPrefixError{prefix}
	}
}

func (prefixNilsFirst) Append(buf []byte, isNil bool) (bool, []byte) {
	return isNil, append(buf, prefixFor(isNil, true))
}

func (prefixNilsFirst) Put(buf []byte, isNil bool) bool {
	buf[0] = prefixFor(isNil, true)
	return isNil
}

func (prefixNilsFirst) Get(buf []byte) bool {
	done, err := eval(buf[0], true)
	if err != nil {
		panic(err)
	}
	return done
}

func (prefixNilsFirst) Write(w io.Writer, isNil bool) (bool, error) {
	if _, err := w.Write([]byte{prefixFor(isNil, true)}); err != nil {
		return true, err
	}
	return isNil, nil
}

func (prefixNilsFirst) Read(r io.Reader) (bool, error) {
	prefix := []byte{0}
	_, err := io.ReadFull(r, prefix)
	if err != nil {
		return true, err
	}
	return eval(prefix[0], true)
}

func (prefixNilsLast) Append(buf []byte, isNil bool) (bool, []byte) {
	return isNil, append(buf, prefixFor(isNil, false))
}

func (prefixNilsLast) Put(buf []byte, isNil bool) bool {
	buf[0] = prefixFor(isNil, false)
	return isNil
}

func (prefixNilsLast) Get(buf []byte) bool {
	done, err := eval(buf[0], false)
	if err != nil {
		panic(err)
	}
	return done
}

func (prefixNilsLast) Write(w io.Writer, isNil bool) (bool, error) {
	if _, err := w.Write([]byte{prefixFor(isNil, false)}); err != nil {
		return true, err
	}
	return isNil, nil
}

func (prefixNilsLast) Read(r io.Reader) (bool, error) {
	prefix := []byte{0}
	_, err := io.ReadFull(r, prefix)
	if err != nil {
		return true, err
	}
	return eval(prefix[0], false)
}
