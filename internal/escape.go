package internal

import (
	"bytes"
	"io"
)

// Functions for delimiting and escaping.
// The lexicographical binary ordering of encoded aggregates is preserved.
// For example, ["ab", "cde"] is less than ["aba", "de"], because "ab" is less than "aba".
// The delimiter can't itself be used to escape a delimiter because it leads to ambiguities,
// so there needs to be a distinct escape character.

// This comment explains why the delimiter and escape values must be 0x00 and 0x01.
// Strings are used for clarity, with "," and "\" denoting the delimiter and escape bytes.
// All input characters have their natural meaning (no delimiters or escapes).
// The encodings for maps and structs will be analogous.
//
//      input slice  -> encoded string
//   A: ["a", "bc"]  -> a,bc
//   B: ["a", ",bc"] -> a,\,bc
//   C: ["a", "\bc"] -> a,\\bc
//   D: ["ab", "c"]  -> ab,c
//   E: ["a,", "bc"] -> a\,,bc
//   F: ["a\", "bc"] -> a\\,bc
//
// B and E are an example of why the delimiter can't be its own escape,
// the encoded strings would both be "a,,,b".
//
// A, B, and C must all be less than D, E, and F.
// We can see "," must be less than all other values including the escape, so it must be 0x00.
//
// Since "," is less than everything else, E < D (first slice element "a," < "ab"). Therefore "a\,,bc" < "ab,c".
// We can see "\" must be less than all other values except the delimiter, so it must be 0x01.

// delimiterByte is used to delimit elements of an aggregate value.
const delimiterByte byte = 0x00

// escapeByte is used the escape the delimiter and escape bytes when they appear in data.
//
// This includes appearing in the encodings of nested aggregates,
// because those are still just data at the level of the enclosing aggregate.
const escapeByte byte = 0x01

// Convenience byte slices.
var (
	del    = []byte{delimiterByte}
	esc    = []byte{escapeByte}
	escDel = []byte{escapeByte, delimiterByte}
	escEsc = []byte{escapeByte, escapeByte}
)

// Escape writes p to w, escaping all delimiters and escapes first.
// Escape does not write an unescaped trailing delimiter.
// It returns the number of bytes read from p.
func Escape(w io.Writer, p []byte) (int, error) {
	var n int // running count of the number of bytes of p successfully processed.
	for i, b := range p {
		switch b {
		case delimiterByte:
			count, err := w.Write(p[n:i])
			n += count
			if err != nil {
				return n, err
			}
			if _, err = w.Write(escDel); err != nil {
				return n, err
			}
			n++
		case escapeByte:
			count, err := w.Write(p[n:i])
			n += count
			if err != nil {
				return n, err
			}
			if _, err = w.Write(escEsc); err != nil {
				return n, err
			}
			n++
		}
	}
	if n < len(p) {
		count, err := w.Write(p[n:])
		n += count
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Unescape reads from r until the first unescaped delimiter or io.EOF,
// returning the unescaped data without the trailing delimiter, if any.
//
// Unescape inherits its error behavior from r.
// In particular, it may return a non-nil error from the same call when encountered,
// or return the error (and no data) from a subsequent call.
// If err is non-nil, the data is valid for what was read from r.
func Unescape(r io.Reader) ([]byte, error) {
	// Reading from r one byte at a time, because we can't unread.
	in := []byte{0}
	var out bytes.Buffer

	var escaped bool // if the previous byte read is an escape
	for {
		n, err := r.Read(in)
		if n == 0 {
			if err != nil {
				return out.Bytes(), err
			}
			// no data read and err == nil is allowed
			continue
		}
		switch in[0] {
		case delimiterByte:
			if !escaped {
				return out.Bytes(), nil
			}
			escaped = false
		case escapeByte:
			if !escaped {
				escaped = true
				continue
			}
		}
		escaped = false
		out.WriteByte(in[0])
	}
}
