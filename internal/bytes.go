package internal

import (
	"bytes"
	"io"
)

// bytesCodec is the Codec for byte slices.
//
// []byte is slightly different than string because it can be nil.
// This is more efficient than sliceCodec would be.
type bytesCodec[S ~[]byte] struct {
	writePrefix prefixWriter[S]
}

func BytesCodec[S ~[]byte](nilsFirst bool) Codec[S] {
	return bytesCodec[S]{getPrefixWriter[S](isNilSlice, nilsFirst)}
}

func (c bytesCodec[S]) Read(r io.Reader) (S, error) {
	if isNil, err := ReadPrefix(r); isNil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	// io.Copy will not return io.EOF
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	return S(buf.Bytes()), nil
}

func (c bytesCodec[S]) Write(w io.Writer, value S) error {
	if done, err := c.writePrefix(w, value); done {
		return err
	}
	_, err := w.Write([]byte(value))
	return err
}

func (c bytesCodec[S]) RequiresTerminator() bool {
	return true
}
