package lexy

import (
	"bytes"
	"errors"
	"io"
)

// terminatorCodec escapes and terminates data written by codec,
// and performs the inverse operation when reading.
//
// Read only reads up to the first unescaped terminator byte,
// which will have been previously written by Write.
// This is the entire point of this Codec, to bound a Read that otherwise would not be bound.
type terminatorCodec[T any] struct {
	codec Codec[T]
}

// Codec for terminating and escaping.
// The lexicographical binary ordering of encoded aggregates is preserved.
// For example, {"ab", "cde"} is less than {"aba", "de"}, because "ab" is less than "aba".
// The terminator can't itself be used to escape a terminator because it leads to ambiguities,
// so there needs to be a distinct escape character.

// This comment explains why the terminator and escape values must be 0x00 and 0x01.
// Strings are used for clarity, with "," and "\" denoting the terminator and escape bytes.
// All input characters have their natural meaning (no terminators or escapes).
// The encodings for maps and structs will be analogous.
//
//	input slice  -> encoded string
//	A: {"a", "bc"}  -> a,bc
//	B: {"a", ",bc"} -> a,\,bc
//	C: {"a", "\bc"} -> a,\\bc
//	D: {"ab", "c"}  -> ab,c
//	E: {"a,", "bc"} -> a\,,bc
//	F: {"a\", "bc"} -> a\\,bc
//
// B and E are an example of why the terminator can't be its own escape,
// the encoded strings would both be "a,,,b".
//
// A, B, and C must all be less than D, E, and F.
// We can see "," must be less than all other values including the escape, so it must be 0x00.
//
// Since "," is less than everything else, E < D (first slice element "a," < "ab"). Therefore "a\,,bc" < "ab,c".
// We can see "\" must be less than all other values except the terminator, so it must be 0x01.
const (
	// terminator is used to terminate elements, when necessary.
	terminator byte = 0x00

	// escape is used the escape the terminator and escape bytes when they appear in data, when necessary.
	// This includes those values appearing in the encodings of nested aggregates,
	// because those are still just data at the level of the enclosing aggregate.
	escape byte = 0x01
)

// Convenience byte slices for writers.
var (
	term    = []byte{terminator}
	escTerm = []byte{escape, terminator}
	escEsc  = []byte{escape, escape}
)

func (c terminatorCodec[T]) Read(r io.Reader) (T, error) {
	var zero T
	b, readErr := doUnescape(r)
	if errors.Is(readErr, io.EOF) && len(b) == 0 {
		return zero, io.EOF
	}
	if readErr != nil {
		// The trailing terminator was not reached, we do not have complete data.
		return zero, UnexpectedIfEOF(readErr)
	}
	value, codecErr := c.codec.Read(bytes.NewReader(b))
	if codecErr != nil {
		return zero, UnexpectedIfEOF(codecErr)
	}
	return value, nil
}

func (c terminatorCodec[T]) Write(w io.Writer, value T) error {
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufSize))
	if err := c.codec.Write(buf, value); err != nil {
		return err
	}
	if _, err := doEscape(w, buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (c terminatorCodec[T]) RequiresTerminator() bool {
	return false
}

// doEscape writes p to w, escaping all terminators and escapes first, and then writes a final terminator.
// It returns the number of bytes read from p.
func doEscape(w io.Writer, p []byte) (int, error) {
	// running count of the number of bytes of p successfully processed
	// also used as the start of the next block of bytes to write
	var n int
	for i, b := range p {
		switch b {
		case terminator:
			count, err := w.Write(p[n:i])
			n += count
			if err != nil {
				return n, err
			}
			if _, err = w.Write(escTerm); err != nil {
				return n, err
			}
			n++
		case escape:
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
	if _, err := w.Write(term); err != nil {
		return n, err
	}
	return n, nil
}

// doUnescape reads and unescapes data from r until the first unescaped terminator,
// or until no bytes are read from r and an error occurs.
// doUnescape does not return the trailing terminator.
// If the returned error is non-nil, the unescaped terminator was not reached.
// However, the data is valid for what was read from r,
// with the possible exception of missing a trailing escape.
//
// doUnescape will continue reading from r if no bytes are read and no error occurs.
// doUnescape will continue reading from r if a byte was read and an error occurs,
// and will ignore the error assuming it will reoccur on the next read.
func doUnescape(r io.Reader) ([]byte, error) {
	// Reading from r one byte at a time, because we can't unread.
	in := []byte{0}
	out := bytes.NewBuffer(make([]byte, 0, defaultBufSize))

	escaped := false // if the previous byte read is an escape
	for {
		_, err := io.ReadFull(r, in)
		if err != nil {
			return out.Bytes(), err
		}
		// handle unescaped terminators and escapes
		// everything else goes into the output as-is
		if !escaped {
			if in[0] == terminator {
				return out.Bytes(), nil
			}
			if in[0] == escape {
				escaped = true
				continue
			}
		}
		escaped = false
		out.WriteByte(in[0])
	}
}
