package lexy

import (
	"encoding/binary"
	"io"
)

// Codecs for bool and fixed-length unsigned integral types.
// These are:
//   - bool
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//
// These encode a value in big-endian order.
type (
	boolCodec   struct{}
	uintCodec   struct{}
	uint8Codec  struct{}
	uint16Codec struct{}
	uint32Codec struct{}
	uint64Codec struct{}
)

func (boolCodec) Read(r io.Reader) (bool, error) {
	var value bool
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return false, err
	}
	return value, nil
}

func (boolCodec) Write(w io.Writer, value bool) error {
	return binary.Write(w, binary.BigEndian, value)
}

func (boolCodec) RequiresTerminator() bool {
	return false
}

func (uintCodec) Read(r io.Reader) (uint, error) {
	value, err := stdUint64.Read(r)
	return uint(value), err
}

func (uintCodec) Write(w io.Writer, value uint) error {
	return stdUint64.Write(w, uint64(value))
}

func (uintCodec) RequiresTerminator() bool {
	return false
}

func (uint8Codec) Read(r io.Reader) (uint8, error) {
	var value uint8
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (uint8Codec) Write(w io.Writer, value uint8) error {
	return binary.Write(w, binary.BigEndian, value)
}

func (uint8Codec) RequiresTerminator() bool {
	return false
}

func (uint16Codec) Read(r io.Reader) (uint16, error) {
	var value uint16
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (uint16Codec) Write(w io.Writer, value uint16) error {
	return binary.Write(w, binary.BigEndian, value)
}

func (uint16Codec) RequiresTerminator() bool {
	return false
}

func (uint32Codec) Read(r io.Reader) (uint32, error) {
	var value uint32
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (uint32Codec) Write(w io.Writer, value uint32) error {
	return binary.Write(w, binary.BigEndian, value)
}

func (uint32Codec) RequiresTerminator() bool {
	return false
}

func (uint64Codec) Read(r io.Reader) (uint64, error) {
	var value uint64
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (uint64Codec) Write(w io.Writer, value uint64) error {
	return binary.Write(w, binary.BigEndian, value)
}

func (uint64Codec) RequiresTerminator() bool {
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

func (asInt64Codec[T]) Read(r io.Reader) (T, error) {
	value, err := stdInt64.Read(r)
	return T(value), err
}

func (asInt64Codec[T]) Write(w io.Writer, value T) error {
	return stdInt64.Write(w, int64(value))
}

func (asInt64Codec[T]) RequiresTerminator() bool {
	return false
}
