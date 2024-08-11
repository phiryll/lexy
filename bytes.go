package lexy

import (
	"bytes"
	"io"
)

// bytesCodec is the Codec for byte slices.
//
// Read will fully consume its argument io.Reader if the value is not nil.
// []byte is slightly different than string because it can be nil.
// This is more efficient than sliceCodec would be.
type bytesCodec[S ~[]byte] struct {
	nilsFirst bool
}

func (c bytesCodec[S]) Read(r io.Reader) (S, error) {
	if done, err := ReadPrefix(r); done {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufSize))
	// io.Copy will not return io.EOF
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return S(buf.Bytes()), nil
}

func (c bytesCodec[S]) Write(w io.Writer, value S) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	_, err := w.Write([]byte(value))
	return err
}

func (c bytesCodec[S]) RequiresTerminator() bool {
	return true
}

func (c bytesCodec[S]) NilsLast() NillableCodec[S] {
	return bytesCodec[S]{false}
}
