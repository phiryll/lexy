package internal

import (
	"bytes"
	"io"
	"math"
)

// Codecs used by other Codecs.
var (
	uint32Codec   Codec[uint32]  = uintCodec[uint32]{}
	uint64Codec   Codec[uint64]  = uintCodec[uint64]{}
	int8Codec     Codec[int8]    = intCodec[int8]{signBit: math.MinInt8}
	int32Codec    Codec[int32]   = intCodec[int32]{signBit: math.MinInt32}
	int64Codec    Codec[int64]   = intCodec[int64]{signBit: math.MinInt64}
	aFloat32Codec Codec[float32] = Float32Codec[float32]()
	aFloat64Codec Codec[float64] = Float64Codec[float64]()
)

// Same interface as lexy.Codec, to avoid a circular dependency.
// lexy.Codec cannot be a type alias to this, because generic type aliases are not permitted.
type Codec[T any] interface {
	Read(io.Reader) (T, error)
	Write(io.Writer, T) error
	RequiresTerminator() bool
}

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write when encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	if err := codec.Write(buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read when decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return codec.Read(bytes.NewReader(data))
}

func unexpectedIfEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
