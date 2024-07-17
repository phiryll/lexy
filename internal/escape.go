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
//	input slice  -> encoded string
//	A: ["a", "bc"]  -> a,bc
//	B: ["a", ",bc"] -> a,\,bc
//	C: ["a", "\bc"] -> a,\\bc
//	D: ["ab", "c"]  -> ab,c
//	E: ["a,", "bc"] -> a\,,bc
//	F: ["a\", "bc"] -> a\\,bc
//
// B and E are an example of why the delimiter can't be its own escape,
// the encoded strings would both be "a,,,b".
//
// A, B, and C must all be less than D, E, and F.
// We can see "," must be less than all other values including the escape, so it must be 0x00.
//
// Since "," is less than everything else, E < D (first slice element "a," < "ab"). Therefore "a\,,bc" < "ab,c".
// We can see "\" must be less than all other values except the delimiter, so it must be 0x01.
const (
	// DelimiterByte is used to delimit elements of an aggregate value.
	DelimiterByte byte = 0x00

	// EscapeByte is used the escape the delimiter and escape bytes when they appear in data.
	//
	// This includes appearing in the encodings of nested aggregates,
	// because those are still just data at the level of the enclosing aggregate.
	EscapeByte byte = 0x01
)

// Convenience byte slices for writers.
var (
	del    = []byte{DelimiterByte}
	escDel = []byte{EscapeByte, DelimiterByte}
	escEsc = []byte{EscapeByte, EscapeByte}
)

// TerminateIfNeeded returns a Codec that uses codec,
// escaping and terminating if codec.RequiresTerminator() is true.
// The returned Codec may not be thread-safe, and MUST be created anew when used.
func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	// This also covers the case if codec is an escaper.
	if !codec.RequiresTerminator() {
		return codec
	}
	return terminator[T]{codec: codec}
}

type terminator[T any] struct {
	codec   Codec[T]
	scratch bytes.Buffer
}

func (c terminator[T]) Read(r io.Reader) (T, error) {
	var value T
	b, readErr := unescape(r)
	if readErr != nil && (readErr != io.EOF || len(b) == 0) {
		return value, readErr
	}
	value, codecErr := c.codec.Read(bytes.NewBuffer(b))
	if codecErr != nil && codecErr != io.EOF {
		return value, codecErr
	}
	return value, nil
}

func (c terminator[T]) Write(w io.Writer, value T) error {
	c.scratch.Reset()
	if err := c.codec.Write(&c.scratch, value); err != nil {
		return err
	}
	if _, err := escape(w, c.scratch.Bytes()); err != nil {
		return err
	}
	return nil
}

func (c terminator[T]) RequiresTerminator() bool {
	return false
}

var ExportEscapeForTesting = escape
var ExportUnescapeForTesting = unescape

// escape writes p to w, escaping all delimiters and escapes first, and writing a final terminator.
// It returns the number of bytes read from p.
func escape(w io.Writer, p []byte) (int, error) {
	// running count of the number of bytes of p successfully processed
	// also used as the start of the next block of bytes to write
	var n int
	for i, b := range p {
		switch b {
		case DelimiterByte:
			count, err := w.Write(p[n:i])
			n += count
			if err != nil {
				return n, err
			}
			if _, err = w.Write(escDel); err != nil {
				return n, err
			}
			n++
		case EscapeByte:
			count, err := w.Write(p[n:i])
			n += count
			if err != nil {
				return n, err
			}
			if _, err = w.Write(escEsc); err != nil {
				return n, err
			}
			n++
		default:
			// do nothing
		}
	}
	if n < len(p) {
		count, err := w.Write(p[n:])
		n += count
		if err != nil {
			return n, err
		}
	}
	if _, err := w.Write(del); err != nil {
		return n, err
	}
	return n, nil
}

// unescape reads from r until the first unescaped delimiter or io.EOF,
// returning the unescaped data without the trailing delimiter, if any.
//
// unescape inherits its error behavior from r.
// In particular, it may return a non-nil error from the same call when encountered,
// or return the error (and no data) from a subsequent call.
// If err is non-nil, the data is valid for what was read from r.
func unescape(r io.Reader) ([]byte, error) {
	// Reading from r one byte at a time, because we can't unread.
	in := []byte{0}
	var out bytes.Buffer

	escaped := false // if the previous byte read is an escape
	for {
		n, err := r.Read(in)
		if n == 0 {
			if err != nil {
				return out.Bytes(), err
			}
			// no data read and err == nil is allowed
			continue
		}
		// handle unescaped delimiters and escapes
		// everything else goes into the output as-is
		if !escaped {
			if in[0] == DelimiterByte {
				return out.Bytes(), nil
			}
			if in[0] == EscapeByte {
				escaped = true
				continue
			}
		}
		escaped = false
		out.WriteByte(in[0])
	}
}
