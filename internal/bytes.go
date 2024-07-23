package internal

import (
	"bytes"
	"io"
)

// bytesCodec is the Codec for byte slices.
// Use MakeBytesCodec[S ~[]byte]() to create a new bytesCodec.
//
// []byte is slightly different than string because it can be nil.
// This is more efficient than sliceCodec would be.
type bytesCodec[S ~[]byte] struct{}

func BytesCodec[S ~[]byte]() Codec[S] {
	return bytesCodec[S]{}
}

func (c bytesCodec[S]) Read(r io.Reader) (S, error) {
	empty := S{}
	if value, done, err := ReadPrefix(r, true, &empty); done {
		return value, err
	}
	var buf bytes.Buffer
	// io.Copy will not return io.EOF
	n, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return S(buf.Bytes()), nil
}

func (c bytesCodec[S]) Write(w io.Writer, value S) error {
	if done, err := WritePrefix(w, isNilSlice, isEmptySlice, value); done {
		return err
	}
	_, err := w.Write([]byte(value))
	return err
}

func (c bytesCodec[S]) RequiresTerminator() bool {
	return true
}
