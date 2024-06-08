package internal

import "io"

type StringCodec struct{}

func (c StringCodec) Read(r io.Reader) (string, error) {
	panic("unimplemented")
}

func (c StringCodec) Write(value string, w io.Writer) error {
	panic("unimplemented")
}
