package lexy

import "bytes"

// terminatorCodec escapes and terminates data written by codec,
// and performs the inverse operation when reading.
//
// Get only reads up to the first unescaped terminator byte (which it consumes but does not return),
// which will have been previously written by Append or Put.
type terminatorCodec[T any] struct {
	codec Codec[T]
}

// The lexicographical binary ordering of encoded aggregates is preserved.
// For example, {"ab", "cde"} is less than {"aba", "de"}, because "ab" is less than "aba".
// The terminator can't itself be used to escape a terminator because it leads to ambiguities,
// so there needs to be a distinct escape character.
//
// The rest of this comment explains why the terminator and escape values must be 0x00 and 0x01.
// Strings are used for clarity, with "," and "\" denoting the terminator and escape bytes.
// All input characters have their natural meaning (no terminators or escapes).
// The encodings for maps and structs will be analogous.
//
//	input slice  -> encoded string
//	A: {"a", "bc"}  -> a,bc
//	B: {"a", ",bc"} -> a,\,bc
//	C: {"a", "\bc"} -> a,\\bc
//	D: {"ab", "c"}  -> ab,c
//	E: {"a,", "bc"} -> a\,,bc
//	F: {"a\", "bc"} -> a\\,bc
//
// B and E are an example of why the terminator can't be its own escape,
// the encoded strings would both be "a,,,b".
//
// A, B, and C must all be less than D, E, and F.
// We can see "," must be less than all other values including the escape, so it must be 0x00.
//
// Since "," is less than everything else, E < D (first slice element "a," < "ab"). Therefore "a\,,bc" < "ab,c".
// We can see "\" must be less than all other values except the terminator, so it must be 0x01.
const (
	// terminator is used to terminate elements, when necessary.
	terminator byte = 0x00

	// escape is used the escape the terminator and escape bytes when they appear in data, when necessary.
	// This includes those values appearing in the encodings of nested aggregates,
	// because those are still just data at the level of the enclosing aggregate.
	escape byte = 0x01
)

func (c terminatorCodec[T]) Append(buf []byte, value T) []byte {
	start := len(buf)
	buf = c.codec.Append(buf, value)
	n := termNumAdded(buf[start:])
	buf = append(buf, make([]byte, n)...)
	term(buf[start:], n)
	return buf
}

func (c terminatorCodec[T]) Put(buf []byte, value T) []byte {
	original := buf
	buf = c.codec.Put(buf, value)
	numPut := len(original) - len(buf)
	n := termNumAdded(original[:numPut])
	term(original[:numPut+n], n)
	return buf[n:]
}

func (c terminatorCodec[T]) Get(buf []byte) (T, []byte) {
	encodedValue, buf := termGet(buf)
	value, _ := c.codec.Get(encodedValue)
	return value, buf
}

func (terminatorCodec[T]) RequiresTerminator() bool {
	return false
}

var (
	eByte = []byte{escape}
	tByte = []byte{terminator}
)

// termNumAdded returns how many more bytes need to be added to escape and terminate buf.
func termNumAdded(buf []byte) int {
	//nolint:mnd
	if len(buf) > 64 {
		// This performs better for larger inputs, on systems with native implementations of bytes.Count.
		return bytes.Count(buf, eByte) + bytes.Count(buf, tByte) + 1
	}
	n := 0
	for _, b := range buf {
		if b == escape || b == terminator {
			n++
		}
	}
	return n + 1 // final terminator
}

// term escapes and terminates buf[:len(buf)-n] in-place, expanding into the last n bytes.
func term(buf []byte, n int) {
	// Going backwards ensures that every byte is copied at most once.
	dst := len(buf) - 1
	buf[dst] = terminator
	dst--
	for i := len(buf) - n - 1; i != dst; i-- {
		buf[dst] = buf[i]
		dst--
		if buf[i] == escape || buf[i] == terminator {
			buf[dst] = escape
			dst--
		}
	}
}

func termAppend(buf, value []byte) []byte {
	buf = extend(buf, len(value))
	for _, b := range value {
		if b == escape || b == terminator {
			buf = append(buf, escape)
		}
		buf = append(buf, b)
	}
	return append(buf, terminator)
}

func termGet(buf []byte) ([]byte, []byte) {
	value := make([]byte, 0, len(buf))
	escaped := false // if the previous byte read is an escape
	for i, b := range buf {
		// handle unescaped terminators and escapes
		// everything else goes into the output as-is
		if !escaped {
			if b == terminator {
				return value, buf[i+1:]
			}
			if b == escape {
				escaped = true
				continue
			}
		}
		escaped = false
		value = append(value, b)
	}
	panic(errUnterminatedBuffer)
}
