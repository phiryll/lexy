package internal

import (
	"io"
	"math/big"
)

var (
	BigIntCodec   Codec[*big.Int]   = bigIntCodec{}
	BigFloatCodec Codec[*big.Float] = bigFloatCodec{}
)

// bigIntCodec is the Codec for big.Int values.
//
// Values are encoded using this logic:
//
//	b := value.Bytes() // absolute value as a big-endian byte slice
//	size := len(b)
//	if value < 0 {
//		write -size using Int64Codec
//		write b with all bits flipped
//	} else {
//		write +size using Int64Codec
//		write b
//	}
//
// This makes size (negative for negative values) the primary sort key,
// and the big-endian bytes for the value (bits flipped for negative values) the secondary sort key.
// The effect is that longer numbers will be ordered closer to +/-infinity.
// This works because bigInt.Bytes() will never have a leading zero byte.
type bigIntCodec struct{}

func (c bigIntCodec) Read(r io.Reader) (*big.Int, error) {
	neg := false
	size, err := Int64Codec.Read(r)
	if err != nil {
		return nil, err
	}
	if size < 0 {
		neg = true
		size = -size
	}
	b := make([]byte, size)
	n, err := r.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err == io.EOF {
		if int64(n) < size {
			return nil, io.ErrUnexpectedEOF
		}
		err = nil
	}
	if neg {
		invertSlice(b)
	}
	var value big.Int
	value.SetBytes(b)
	if neg {
		value.Neg(&value)
	}
	return &value, nil
}

func (c bigIntCodec) Write(w io.Writer, value *big.Int) error {
	sign := value.Sign()
	b := value.Bytes()
	size := len(b)
	if sign < 0 {
		size = -size
		invertSlice(b)
	}
	if err := Int64Codec.Write(w, int64(size)); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

type bigFloatCodec struct{}

func (c bigFloatCodec) Read(r io.Reader) (*big.Float, error) {
	panic("unimplemented")
}

func (c bigFloatCodec) Write(w io.Writer, value *big.Float) error {
	panic("unimplemented")
}
