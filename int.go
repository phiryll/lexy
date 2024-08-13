package lexy

import (
	"encoding/binary"
	"io"
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

func (c uintCodec[T]) RequiresTerminator() bool {
	return false
}

type asUint64Codec[T ~uint] struct{}

func (c asUint64Codec[T]) Read(r io.Reader) (T, error) {
	value, err := stdUint64.Read(r)
	return T(value), err
}

func (c asUint64Codec[T]) Write(w io.Writer, value T) error {
	return stdUint64.Write(w, uint64(value))
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
	value, err := stdInt64.Read(r)
	return T(value), err
}

func (c asInt64Codec[T]) Write(w io.Writer, value T) error {
	return stdInt64.Write(w, int64(value))
}

func (c asInt64Codec[T]) RequiresTerminator() bool {
	return false
}
