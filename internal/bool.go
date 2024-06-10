package internal

import (
	"fmt"
	"io"
)

type BoolCodec struct{}

func (c BoolCodec) Read(r io.Reader) (bool, error) {
	in := make([]byte, 1)
	if _, err := r.Read(in); err != nil {
		return false, err
	}
	if in[0] != 0 && in[0] != 1 {
		return false, fmt.Errorf("value should be 0 or 1, but was: %d", in[0])
	}
	return in[0] == 1, nil
}

func (c BoolCodec) Write(value bool, w io.Writer) error {
	var out byte
	if value {
		out = 1
	}
	_, err := w.Write([]byte{out})
	return err
}
