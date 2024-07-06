package internal

import (
	"io"
	"math/big"
)

var (
	BigIntCodec   Codec[big.Int]   = bigIntCodec{}
	BigFloatCodec Codec[big.Float] = bigFloatCodec{}
)

type bigIntCodec struct{}

func (c bigIntCodec) Read(r io.Reader) (big.Int, error) {
	panic("unimplemented")
}

func (c bigIntCodec) Write(w io.Writer, value big.Int) error {
	panic("unimplemented")
}

type bigFloatCodec struct{}

func (c bigFloatCodec) Read(r io.Reader) (big.Float, error) {
	panic("unimplemented")
}

func (c bigFloatCodec) Write(w io.Writer, value big.Float) error {
	panic("unimplemented")
}
