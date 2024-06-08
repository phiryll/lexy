package internal

import (
	"io"
	"math/big"
)

type BigIntCodec struct{}

func (c BigIntCodec) Read(r io.Reader) (big.Int, error) {
	panic("unimplemented")
}

func (c BigIntCodec) Write(value big.Int, w io.Writer) error {
	panic("unimplemented")
}

type BigFloatCodec struct{}

func (c BigFloatCodec) Read(r io.Reader) (big.Float, error) {
	panic("unimplemented")
}

func (c BigFloatCodec) Write(value big.Float, w io.Writer) error {
	panic("unimplemented")
}
