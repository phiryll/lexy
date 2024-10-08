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
	//	func (fooCodec) Append(buf []byte, value Foo) []byte {
	//	    done, buf := PrefixNilsFirst.Append(buf, value == nil)
	//	    if done {
	//	        return buf
	//	    }
	//	    // encode and append the non-nil value to buf
	//	}
	Append(buf []byte, isNil bool) (done bool, newBuf []byte)

	// Put sets buf[0] to a prefix byte.
	// This is a typical usage:
	//
	//	func (fooCodec) Put(buf []byte, value Foo) []byte {
	//	    done, buf := PrefixNilsFirst.Put(buf, value == nil)
	//	    if done {
	//	        return buf
	//	    }
	//	    // encode the non-nil value into buf
	//	}
	Put(buf []byte, isNil bool) (done bool, newBuf []byte)

	// Get decodes a prefix byte from buf[0].
	// Get will panic if the prefix byte is invalid.
	// Get will not modify buf.
	// This is a typical usage:
	//
	//	func (c fooCodec) Get(buf []byte) (Foo, []byte)
	//	    done, buf := PrefixNilsFirst.Get(buf)
	//	    if done {
	//	        return nil, buf
	//	    }
	//	    // decode and return a non-nil value from buf
	//	}
	Get(buf []byte) (done bool, newBuf []byte)
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
func (prefixNilsFirst) Append(buf []byte, isNil bool) (bool, []byte) {
	if isNil {
		return true, append(buf, prefixNilFirst)
	}
	return false, append(buf, prefixNonNil)
}

//nolint:revive
func (prefixNilsFirst) Put(buf []byte, isNil bool) (bool, []byte) {
	if isNil {
		buf[0] = prefixNilFirst
		return true, buf[1:]
	}
	buf[0] = prefixNonNil
	return false, buf[1:]
}

func (prefixNilsFirst) Get(buf []byte) (bool, []byte) {
	switch buf[0] {
	case prefixNonNil:
		return false, buf[1:]
	case prefixNilFirst:
		return true, buf[1:]
	case prefixNilLast:
		panic(errUnexpectedNilsLast)
	default:
		panic(unknownPrefixError{buf[0]})
	}
}

//nolint:revive
func (prefixNilsLast) Append(buf []byte, isNil bool) (bool, []byte) {
	if isNil {
		return true, append(buf, prefixNilLast)
	}
	return false, append(buf, prefixNonNil)
}

//nolint:revive
func (prefixNilsLast) Put(buf []byte, isNil bool) (bool, []byte) {
	if isNil {
		buf[0] = prefixNilLast
		return true, buf[1:]
	}
	buf[0] = prefixNonNil
	return false, buf[1:]
}

func (prefixNilsLast) Get(buf []byte) (bool, []byte) {
	switch buf[0] {
	case prefixNonNil:
		return false, buf[1:]
	case prefixNilFirst:
		panic(errUnexpectedNilsFirst)
	case prefixNilLast:
		return true, buf[1:]
	default:
		panic(unknownPrefixError{buf[0]})
	}
}
