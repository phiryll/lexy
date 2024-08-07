package internal

import (
	"encoding/binary"
	"io"
	"math"
)

func BoolCodec[T ~bool]() Codec[T] {
	return uintCodec[T]{}
}

func Uint8Codec[T ~uint8]() Codec[T] {
	return uintCodec[T]{}
}

func Uint16Codec[T ~uint16]() Codec[T] {
	return uintCodec[T]{}
}

func Uint32Codec[T ~uint32]() Codec[T] {
	return uintCodec[T]{}
}

func Uint64Codec[T ~uint64]() Codec[T] {
	return uintCodec[T]{}
}

func UintCodec[T ~uint]() Codec[T] {
	return asUint64Codec[T]{}
}

func Int8Codec[T ~int8]() Codec[T] {
	return intCodec[T]{signBit: math.MinInt8}
}

func Int16Codec[T ~int16]() Codec[T] {
	return intCodec[T]{signBit: math.MinInt16}
}

func Int32Codec[T ~int32]() Codec[T] {
	return intCodec[T]{signBit: math.MinInt32}
}

func Int64Codec[T ~int64]() Codec[T] {
	return intCodec[T]{signBit: math.MinInt64}
}

func IntCodec[T ~int]() Codec[T] {
	return asInt64Codec[T]{}
}

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

func (c uintCodec[T]) RequiresTerminator() bool {
	return false
}

type asUint64Codec[T ~uint] struct{}

func (c asUint64Codec[T]) Read(r io.Reader) (T, error) {
	value, err := uint64Codec.Read(r)
	return T(value), err
}

func (c asUint64Codec[T]) Write(w io.Writer, value T) error {
	return uint64Codec.Write(w, uint64(value))
}

func (c asUint64Codec[T]) RequiresTerminator() bool {
	return false
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
// That this works can be seen from the following signed int -> encoded table.
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

func (c intCodec[T]) RequiresTerminator() bool {
	return false
}

type asInt64Codec[T ~int] struct{}

func (c asInt64Codec[T]) Read(r io.Reader) (T, error) {
	value, err := int64Codec.Read(r)
	return T(value), err
}

func (c asInt64Codec[T]) Write(w io.Writer, value T) error {
	return int64Codec.Write(w, int64(value))
}

func (c asInt64Codec[T]) RequiresTerminator() bool {
	return false
}
