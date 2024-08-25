package lexy

// stringCodec is the Codec for strings.
//
// A string is encoded as its bytes.
// Get will fully consume its argument buffer, and will never return a negative byte count.
//
// The order of strings, and this encoding, may be surprising.
// A string in Go is essentially an immutable []byte without text semantics.
// For an encoded UTF-8 string, the order is the same as the lexicographical order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent. Any ordering could be incorrect in another locale.
type stringCodec struct{}

func (stringCodec) Append(buf []byte, value string) []byte {
	return append(buf, value...)
}

func (stringCodec) Put(buf []byte, value string) int {
	return copyAll(buf, []byte(value))
}

func (stringCodec) Get(buf []byte) (string, []byte) {
	return string(buf), buf[len(buf):]
}

func (stringCodec) RequiresTerminator() bool {
	return true
}
