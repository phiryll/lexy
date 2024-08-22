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
	prefix Prefix
}

func (c bytesCodec) Append(buf, value []byte) []byte {
	return AppendUsingWrite[[]byte](c, buf, value)
}

func (c bytesCodec) Put(buf, value []byte) int {
	return PutUsingAppend[[]byte](c, buf, value)
}

func (c bytesCodec) Get(buf []byte) ([]byte, int) {
	return GetUsingRead[[]byte](c, buf)
}

func (c bytesCodec) Write(w io.Writer, value []byte) error {
	if done, err := c.prefix.Write(w, value == nil); done {
		return err
	}
	_, err := w.Write(value)
	return err
}

func (c bytesCodec) Read(r io.Reader) ([]byte, error) {
	if done, err := c.prefix.Read(r); done {
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

//lint:ignore U1000 this is actually used
func (bytesCodec) nilsLast() Codec[[]byte] {
	return bytesCodec{PrefixNilsLast}
}
