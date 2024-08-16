package lexy

import (
	"bytes"
	"io"
)

// bytesCodec is the Codec for []byte.
//
// Read will fully consume its argument io.Reader if the value is not nil.
// []byte is slightly different than string because it can be nil.
// This is more efficient than sliceCodec would be.
type bytesCodec struct {
	nilsFirst bool
}

func (c bytesCodec) Write(w io.Writer, value []byte) error {
	if done, err := WritePrefix(w, value == nil, c.nilsFirst); done {
		return err
	}
	_, err := w.Write(value)
	return err
}

func (bytesCodec) Read(r io.Reader) ([]byte, error) {
	if done, err := ReadPrefix(r); done {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufSize))
	// io.Copy will not return io.EOF
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (bytesCodec) RequiresTerminator() bool {
	return true
}

func (bytesCodec) NilsLast() NillableCodec[[]byte] {
	return bytesCodec{false}
}
