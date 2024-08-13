package lexy

import (
	"encoding/binary"
	"io"
	"math"
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

// Codecs for fixed-length signed integral types.
// These are:
//   - int8
//   - int16
//   - int32
//   - int64
//
// These encode a value by flipping the sign bit and writing in big-endian order.
// That this works can be seen from the following signed int -> encoded table.
//
//	0x8000... -> 0x0000...  most negative
//	0xFFFF... -> 0x7FFF...  -1
//	0x0000... -> 0x8000...  0
//	0x0000..1 -> 0x8000..1  1
//	0x7FFF... -> 0xFFFF...  most positive
type (
	intCodec   struct{}
	int8Codec  struct{}
	int16Codec struct{}
	int32Codec struct{}
	int64Codec struct{}
)

func (intCodec) Read(r io.Reader) (int, error) {
	value, err := stdInt64.Read(r)
	return int(value), err
}

func (intCodec) Write(w io.Writer, value int) error {
	return stdInt64.Write(w, int64(value))
}

func (intCodec) RequiresTerminator() bool {
	return false
}

func (int8Codec) Read(r io.Reader) (int8, error) {
	var value int8
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return math.MinInt8 ^ value, nil
}

func (int8Codec) Write(w io.Writer, value int8) error {
	return binary.Write(w, binary.BigEndian, math.MinInt8^value)
}

func (int8Codec) RequiresTerminator() bool {
	return false
}

func (int16Codec) Read(r io.Reader) (int16, error) {
	var value int16
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return math.MinInt16 ^ value, nil
}

func (int16Codec) Write(w io.Writer, value int16) error {
	return binary.Write(w, binary.BigEndian, math.MinInt16^value)
}

func (int16Codec) RequiresTerminator() bool {
	return false
}

func (int32Codec) Read(r io.Reader) (int32, error) {
	var value int32
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return math.MinInt32 ^ value, nil
}

func (int32Codec) Write(w io.Writer, value int32) error {
	return binary.Write(w, binary.BigEndian, math.MinInt32^value)
}

func (int32Codec) RequiresTerminator() bool {
	return false
}

func (int64Codec) Read(r io.Reader) (int64, error) {
	var value int64
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return math.MinInt64 ^ value, nil
}

func (int64Codec) Write(w io.Writer, value int64) error {
	return binary.Write(w, binary.BigEndian, math.MinInt64^value)
}

func (int64Codec) RequiresTerminator() bool {
	return false
}
