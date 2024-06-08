package internal

import "io"

type Float32Codec struct{}

func (c Float32Codec) Read(r io.Reader) (float32, error) {
	panic("unimplemented")
}

func (c Float32Codec) Write(value float32, w io.Writer) error {
	panic("unimplemented")
}

type Float64Codec struct{}

func (c Float64Codec) Read(r io.Reader) (float64, error) {
	panic("unimplemented")
}

func (c Float64Codec) Write(value float64, w io.Writer) error {
	panic("unimplemented")
}
