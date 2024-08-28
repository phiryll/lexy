package lexy

// terminatorCodec escapes and terminates data written by codec,
// and performs the inverse operation when reading.
//
// Get only reads up to the first unescaped terminator byte (which it consumes but does not return),
// which will have been previously written by Append or Put.
type terminatorCodec[T any] struct {
	codec Codec[T]
}

// Codec for terminating and escaping.
// The lexicographical binary ordering of encoded aggregates is preserved.
// For example, {"ab", "cde"} is less than {"aba", "de"}, because "ab" is less than "aba".
// The terminator can't itself be used to escape a terminator because it leads to ambiguities,
// so there needs to be a distinct escape character.

// This comment explains why the terminator and escape values must be 0x00 and 0x01.
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
	return escapeAppend(buf, c.codec.Append(nil, value))
}

func (c terminatorCodec[T]) Put(buf []byte, value T) []byte {
	return escapePut(buf, c.codec.Append(nil, value))
}

func (c terminatorCodec[T]) Get(buf []byte) (T, []byte) {
	b, buf, _ := unescape(buf)
	value, _ := c.codec.Get(b)
	return value, buf
}

func (terminatorCodec[T]) RequiresTerminator() bool {
	return false
}

func escapeAppend(buf, value []byte) []byte {
	buf = extend(buf, len(value))
	for _, b := range value {
		switch b {
		case terminator:
			buf = append(buf, escape, terminator)
		case escape:
			buf = append(buf, escape, escape)
		default:
			buf = append(buf, b)
		}
	}
	return append(buf, terminator)
}

func escapePut(buf, value []byte) []byte {
	i := 0
	for _, b := range value {
		if b == escape || b == terminator {
			buf[i] = escape
			i++
		}
		buf[i] = b
		i++
	}
	buf[i] = terminator
	return buf[i+1:]
}

// unescape reads and unescapes data from buf until the first unescaped terminator,
// returning the unescaped data, the following buf, and number of bytes read from buf.
// unescape will panic if no unescaped terminator is found.
//
//nolint:nonamedreturns
func unescape(buf []byte) (unescaped, newBuf []byte, numRead int) {
	out := make([]byte, 0, len(buf))
	escaped := false // if the previous byte read is an escape
	for i, b := range buf {
		// handle unescaped terminators and escapes
		// everything else goes into the output as-is
		if !escaped {
			if b == terminator {
				return out, buf[i+1:], i + 1
			}
			if b == escape {
				escaped = true
				continue
			}
		}
		escaped = false
		out = append(out, b)
	}
	panic(errUnterminatedBuffer)
}
