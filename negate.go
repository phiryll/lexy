package lexy

// negateCodec negates codec which does not require escaping, reversing the ordering of its encoding.
//
// This Codec simply flips all the encoded bits.
type negateCodec[T any] struct {
	codec Codec[T]
}

// Negate negates buf, in the sense of lexicographical ordering, returning buf.
//
//nolint:unparam  // For some reason, this method is faster if it returns something.
func negate(buf []byte) []byte {
	for i := range buf {
		buf[i] ^= 0xFF
	}
	return buf
}

// negCopy returns a negated copy of buf.
func negCopy(buf []byte) []byte {
	dst := make([]byte, len(buf))
	for i := range buf {
		dst[i] = ^buf[i]
	}
	return dst
}

func (c negateCodec[T]) Append(buf []byte, value T) []byte {
	start := len(buf)
	buf = c.codec.Append(buf, value)
	negate(buf[start:])
	return buf
}

func (c negateCodec[T]) Put(buf []byte, value T) []byte {
	original := buf
	buf = c.codec.Put(buf, value)
	negate(original[:len(original)-len(buf)])
	return buf
}

func (c negateCodec[T]) Get(buf []byte) (T, []byte) {
	value, temp := c.codec.Get(negCopy(buf))
	return value, buf[len(buf)-len(temp):]
}

func (negateCodec[T]) RequiresTerminator() bool {
	return false
}

// negateEscapeCodec negates codec which requires escaping, reversing the ordering of its encoding.
//
// Every encoding will be greater than any prefix of that encoding (definition of lexicographical ordering).
// For example, consider these encodings:
//
//	A = {0x00, 0x02, 0x03}
//	B = {0x00, 0x02, 0x03, 0x00}
//	A < B
//
// This Codec must effectively reverse that for what the delegate codec produces.
// Just flipping all the bits works except when one encoding is the prefix of another.
// The above example with all bits flipped is:
//
//	^A = {0xFF, 0xFD, 0xFC}
//	^B = {0xFF, 0xFD, 0xFC, 0xFF}
//
// We need to transform these results so that -B is less than -A.
// Adding a 0xFF terminator accomplishes this,
// but then we have another escape/terminator problem, just with 0xFF and 0xFE instead of 0x00 and 0x01.
// We can achieve the same effect by always escaping and terminating the normal way,
// and then flip all the bits, inluding the trailing terminator.
// If we do that for the above example, we get the correctly negated ordering.
//
//	esc+term(A) = {0x01, 0x00, 0x02, 0x03, 0x00}
//	esc+term(B) = {0x01, 0x00, 0x02, 0x03, 0x01, 0x00, 0x00}
//
//	^esc+term(A) = {0xFE, 0xFF, 0xFD, 0xFC, 0xFF}
//	^esc+term(B) = {0xFE, 0xFF, 0xFD, 0xFC, 0xFE, 0xFF, 0xFF}
type negateEscapeCodec[T any] struct {
	codec Codec[T]
}

func (c negateEscapeCodec[T]) Append(buf []byte, value T) []byte {
	start := len(buf)
	buf = c.codec.Append(buf, value)
	n := termNumAdded(buf[start:])
	buf = append(buf, make([]byte, n)...)
	negTerm(buf[start:], n)
	return buf
}

func (c negateEscapeCodec[T]) Put(buf []byte, value T) []byte {
	original := buf
	buf = c.codec.Put(buf, value)
	numPut := len(original) - len(buf)
	n := termNumAdded(original[:numPut])
	negTerm(original[:numPut+n], n)
	return buf[n:]
}

func (c negateEscapeCodec[T]) Get(buf []byte) (T, []byte) {
	encodedValue, buf := negTermGet(buf)
	value, _ := c.codec.Get(encodedValue)
	return value, buf
}

func (negateEscapeCodec[T]) RequiresTerminator() bool {
	return false
}

// negTerm is exactly the same as term, except that it negates every byte written.
func negTerm(buf []byte, n int) {
	// Going backwards ensures that every byte is copied at most once.
	dst := len(buf) - 1
	buf[dst] = ^terminator
	dst--
	for i := len(buf) - n - 1; i != dst; i-- {
		buf[dst] = ^buf[i]
		dst--
		if buf[i] == escape || buf[i] == terminator {
			buf[dst] = ^escape
			dst--
		}
	}
	// still need to negate the first block
	for i := dst; i >= 0; i-- {
		buf[i] = ^buf[i]
	}
}

// negTermGet is exactly the same as termGet, except that it negates every byte read first.
func negTermGet(buf []byte) ([]byte, []byte) {
	value := make([]byte, 0, len(buf))
	escaped := false // if the previous byte read is an escape
	for i, b := range buf {
		b = ^b
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
