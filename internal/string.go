package internal

import (
	"fmt"
	"io"
	"strings"
)

// StringCodec is the Codec for strings.
//
// A string is encoded as its bytes following PrefixZeroValue or PrefixNonZeroValue.
// Read will fully consume its argument io.Reader.
// If a string is part of a larger aggregate and not fixed-length,
// its encoding should be escaped and delimiter-terminated by the enclosing Codec.
//
// The order of strings, and this encoding, may be surprising.
// A string in go is essentially an immutable []byte without text semantics.
// If your string is UTF-8, then the order is the same as the order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent. Any order you choose could be incorrect in another locale.
type StringCodec struct{}

func (c StringCodec) Read(r io.Reader) (string, error) {
	// TODO: clean this up a little bit, it's messy
	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		if err == nil {
			return "", fmt.Errorf("no bytes read and no error")
		}
		return "", err
	}
	if prefix[0] == PrefixZeroValue {
		if err == io.EOF {
			return "", nil
		}
		return "", err
	}
	if err != nil {
		return "", err
	}
	if prefix[0] != PrefixNonZeroValue {
		return "", fmt.Errorf("unexpected prefix")
	}
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c StringCodec) Write(w io.Writer, value string) error {
	if value == "" {
		_, err := w.Write(prefixZero)
		return err
	}
	if _, err := w.Write(prefixNonZero); err != nil {
		return err
	}
	_, err := io.WriteString(w, value)
	return err
}
