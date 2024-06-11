package internal

import "io"

type Float32Codec struct{}

func (c Float32Codec) Read(r io.Reader) (float32, error) {
	panic("unimplemented")
}

func (c Float32Codec) Write(w io.Writer, value float32) error {
	panic("unimplemented")
}

type Float64Codec struct{}

func (c Float64Codec) Read(r io.Reader) (float64, error) {
	panic("unimplemented")
}

func (c Float64Codec) Write(w io.Writer, value float64) error {
	panic("unimplemented")
}
