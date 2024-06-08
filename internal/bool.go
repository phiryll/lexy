package internal

import "io"

type BoolCodec struct{}

func (c BoolCodec) Read(r io.Reader) (bool, error) {
	panic("unimplemented")
}

func (c BoolCodec) Write(value bool, w io.Writer) error {
	panic("unimplemented")
}
