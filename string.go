package lexy

import (
	"io"
	"strings"
)

// stringCodec is the Codec for strings.
//
// A string is encoded as its bytes.
// Read will fully consume its argument io.Reader, and will not return io.EOF.
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
	if len(buf) < len(value) {
		panic("buffer is too small")
	}
	copy(buf, value)
	return len(value)
}

func (stringCodec) Write(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}

func (stringCodec) Get(buf []byte) (string, int) {
	return string(buf), len(buf)
}

func (stringCodec) Read(r io.Reader) (string, error) {
	var buf strings.Builder
	// io.Copy will not return io.EOF
	_, err := io.Copy(&buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (stringCodec) MaxSize() int {
	return -1
}

func (stringCodec) RequiresTerminator() bool {
	return true
}
