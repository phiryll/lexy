package internal

import (
	"io"
	"strings"
)

func StringCodec[T ~string]() Codec[T] {
	return stringCodec[T]{}
}

// stringCodec is the Codec for strings.
//
// A string is encoded as its bytes following PrefixEmpty or PrefixNonEmpty.
// Read will fully consume its argument io.Reader if the string is non-empty.
// If a string is part of a larger aggregate and not fixed-length,
// its encoding should be escaped and terminated by the enclosing Codec.
//
// The order of strings, and this encoding, may be surprising.
// A string in go is essentially an immutable []byte without text semantics.
// For an encoded UTF-8 string, the order is the same as the lexicographical order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent. Any ordering could be incorrect in another locale.
type stringCodec[T ~string] struct{}

func (c stringCodec[T]) Read(r io.Reader) (T, error) {
	if value, done, err := ReadPrefix[string](r, false, nil); done {
		return T(value), err
	}
	var buf strings.Builder
	// io.Copy will not return io.EOF
	n, err := io.Copy(&buf, r)
	if err != nil {
		return T(""), err
	}
	if n == 0 {
		return T(""), io.ErrUnexpectedEOF
	}
	return T(buf.String()), nil
}

func (c stringCodec[T]) Write(w io.Writer, value T) error {
	if done, err := WritePrefix(w, nil, isEmptyString, value); done {
		return err
	}
	_, err := io.WriteString(w, string(value))
	return err
}

func (c stringCodec[T]) RequiresTerminator() bool {
	return true
}
