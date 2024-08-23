package lexy

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
// The one exception is Get, which should check for an empty buffer argument first.
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

	// prefixFor returns which prefix byte to write.
	// This method is used by Append and Put.
	prefixFor(isNil bool) byte
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

//nolint:revive
func (prefixNilsFirst) prefixFor(isNil bool) byte {
	if isNil {
		return prefixNilFirst
	}
	return prefixNonNil
}

func (p prefixNilsFirst) Append(buf []byte, isNil bool) (bool, []byte) {
	return isNil, append(buf, p.prefixFor(isNil))
}

func (p prefixNilsFirst) Put(buf []byte, isNil bool) bool {
	buf[0] = p.prefixFor(isNil)
	return isNil
}

func (prefixNilsFirst) Get(buf []byte) bool {
	switch buf[0] {
	case prefixNonNil:
		return false
	case prefixNilFirst:
		return true
	case prefixNilLast:
		panic(errUnexpectedNilsLast)
	default:
		panic(unknownPrefixError{buf[0]})
	}
}

//nolint:revive
func (prefixNilsLast) prefixFor(isNil bool) byte {
	if isNil {
		return prefixNilLast
	}
	return prefixNonNil
}

func (p prefixNilsLast) Append(buf []byte, isNil bool) (bool, []byte) {
	return isNil, append(buf, p.prefixFor(isNil))
}

func (p prefixNilsLast) Put(buf []byte, isNil bool) bool {
	buf[0] = p.prefixFor(isNil)
	return isNil
}

func (prefixNilsLast) Get(buf []byte) bool {
	switch buf[0] {
	case prefixNonNil:
		return false
	case prefixNilFirst:
		panic(errUnexpectedNilsFirst)
	case prefixNilLast:
		return true
	default:
		panic(unknownPrefixError{buf[0]})
	}
}
