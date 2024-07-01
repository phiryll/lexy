package internal

import (
	"fmt"
	"io"
	"strings"
)

// StringCodec is the Codec for strings.
//
// A string is encoded as its bytes following PrefixEmpty or PrefixNonEmpty.
// Read will fully consume its argument io.Reader if the string is non-empty.
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
	prefix := []byte{0}
	n, err := r.Read(prefix)
	if n == 0 {
		if err == nil || err == io.EOF {
			return "", io.ErrUnexpectedEOF
		}
		return "", err
	}
	switch prefix[0] {
	case PrefixEmpty:
		if err != nil && err != io.EOF {
			return "", err
		}
		return "", nil
	case PrefixNonEmpty:
		var buf strings.Builder
		// io.Copy will not return io.EOF
		n, err := io.Copy(&buf, r)
		if err != nil {
			return "", err
		}
		if n == 0 {
			return "", io.ErrUnexpectedEOF
		}
		return buf.String(), nil
	default:
		if err == nil || err == io.EOF {
			err = fmt.Errorf("unexpected prefix %X", prefix[0])
		}
		return "", err
	}
}

func (c StringCodec) Write(w io.Writer, value string) error {
	if value == "" {
		_, err := w.Write(prefixEmpty)
		return err
	}
	if _, err := w.Write(prefixNonEmpty); err != nil {
		return err
	}
	_, err := io.WriteString(w, value)
	return err
}
