package internal

import (
	"io"
	"math/big"
)

type BigIntCodec struct{}

func (c BigIntCodec) Read(r io.Reader) (big.Int, error) {
	panic("unimplemented")
}

func (c BigIntCodec) Write(w io.Writer, value big.Int) error {
	panic("unimplemented")
}

type BigFloatCodec struct{}

func (c BigFloatCodec) Read(r io.Reader) (big.Float, error) {
	panic("unimplemented")
}

func (c BigFloatCodec) Write(w io.Writer, value big.Float) error {
	panic("unimplemented")
}
