package internal

import (
	"io"
	"strings"
)

// StringCodec is the Codec for strings.
//
// A string is encoded as its bytes. Nothing is written for an empty string.
// Read will fully consume its argument io.Reader.
// If a string is part of a larger aggregate and not fixed-length,
// it should be escaped and delimiter-terminated by the enclosing Codec.
//
// The order of strings, and this encoding, may be surprising.
// A string in go is essentially an immutable []byte without text semantics.
// If your string is UTF-8, then the order is the same as the order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent. Any order you choose could be incorrect in another locale.
type StringCodec struct{}

func (c StringCodec) Read(r io.Reader) (string, error) {
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c StringCodec) Write(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}
