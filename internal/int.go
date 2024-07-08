package internal

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

var (
	BoolCodec     Codec[bool]          = uintCodec[bool]{}
	Uint8Codec    Codec[uint8]         = uintCodec[uint8]{}
	Uint16Codec   Codec[uint16]        = uintCodec[uint16]{}
	Uint32Codec   Codec[uint32]        = uintCodec[uint32]{}
	Uint64Codec   Codec[uint64]        = uintCodec[uint64]{}
	Int8Codec     Codec[int8]          = intCodec[int8]{signBit: math.MinInt8}
	Int16Codec    Codec[int16]         = intCodec[int16]{signBit: math.MinInt16}
	Int32Codec    Codec[int32]         = intCodec[int32]{signBit: math.MinInt32}
	Int64Codec    Codec[int64]         = intCodec[int64]{signBit: math.MinInt64}
	DurationCodec Codec[time.Duration] = intCodec[time.Duration]{signBit: math.MinInt64}
)

// uintCodec is the Codec for bool and fixed-length unsigned integral types.
//
// These are:
//   - bool
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//
// This encodes a value in big-endian order.
type uintCodec[T ~bool | ~uint8 | ~uint16 | ~uint32 | ~uint64] struct{}

func (c uintCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		var zero T
		return zero, err
	}
	return value, nil
}

func (c uintCodec[T]) Write(w io.Writer, value T) error {
	return binary.Write(w, binary.BigEndian, value)
}

// intCodec is the Codec for fixed-length signed integral types.
//
// These are:
//   - int8
//   - int16
//   - int32
//   - int64
//
// This encodes a value by flipping the sign bit and writing in big-endian order.
// You can see that this works from the following signed int -> encoded table.
//
//	0x8000... -> 0x0000...  most negative
//	0xFFFF... -> 0x7FFF...  -1
//	0x0000... -> 0x8000...  0
//	0x0000..1 -> 0x8000..1  1
//	0x7FFF... -> 0xFFFF...  most positive
type intCodec[T ~int8 | ~int16 | ~int32 | ~int64] struct {
	signBit T
}

func (c intCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		var zero T
		return zero, err
	}
	return c.signBit ^ value, nil
}

func (c intCodec[T]) Write(w io.Writer, value T) error {
	return binary.Write(w, binary.BigEndian, c.signBit^value)
}
