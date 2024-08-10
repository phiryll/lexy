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
type stringCodec[T ~string] struct{}

func (c stringCodec[T]) Read(r io.Reader) (T, error) {
	var buf strings.Builder
	// io.Copy will not return io.EOF
	_, err := io.Copy(&buf, r)
	if err != nil {
		return T(""), err
	}
	return T(buf.String()), nil
}

func (c stringCodec[T]) Write(w io.Writer, value T) error {
	_, err := io.WriteString(w, string(value))
	return err
}

func (c stringCodec[T]) RequiresTerminator() bool {
	return true
}
