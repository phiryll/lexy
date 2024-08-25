package lexy

import (
	"encoding/binary"
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
	uint8Codec  struct{}
	uint16Codec struct{}
	uint32Codec struct{}
	uint64Codec struct{}
)

const (
	sizeUint8  = 1
	sizeUint16 = 2
	sizeUint32 = 4
	sizeUint64 = 8
)

//nolint:revive
func (boolCodec) Append(buf []byte, value bool) []byte {
	if value {
		return append(buf, 1)
	}
	return append(buf, 0)
}

//nolint:revive
func (boolCodec) Put(buf []byte, value bool) int {
	if value {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	return sizeUint8
}

func (boolCodec) Get(buf []byte) (bool, []byte) {
	return buf[0] != 0, buf[sizeUint8:]
}

func (boolCodec) RequiresTerminator() bool {
	return false
}

func (uint8Codec) Append(buf []byte, value uint8) []byte {
	return append(buf, value)
}

func (uint8Codec) Put(buf []byte, value uint8) int {
	buf[0] = value
	return sizeUint8
}

func (uint8Codec) Get(buf []byte) (uint8, []byte) {
	return buf[0], buf[sizeUint8:]
}

func (uint8Codec) RequiresTerminator() bool {
	return false
}

func (uint16Codec) Append(buf []byte, value uint16) []byte {
	return binary.BigEndian.AppendUint16(buf, value)
}

func (uint16Codec) Put(buf []byte, value uint16) int {
	binary.BigEndian.PutUint16(buf, value)
	return sizeUint16
}

func (uint16Codec) Get(buf []byte) (uint16, []byte) {
	return binary.BigEndian.Uint16(buf), buf[sizeUint16:]
}

func (uint16Codec) RequiresTerminator() bool {
	return false
}

func (uint32Codec) Append(buf []byte, value uint32) []byte {
	return binary.BigEndian.AppendUint32(buf, value)
}

func (uint32Codec) Put(buf []byte, value uint32) int {
	binary.BigEndian.PutUint32(buf, value)
	return sizeUint32
}

func (uint32Codec) Get(buf []byte) (uint32, []byte) {
	return binary.BigEndian.Uint32(buf), buf[sizeUint32:]
}

func (uint32Codec) RequiresTerminator() bool {
	return false
}

func (uint64Codec) Append(buf []byte, value uint64) []byte {
	return binary.BigEndian.AppendUint64(buf, value)
}

func (uint64Codec) Put(buf []byte, value uint64) int {
	binary.BigEndian.PutUint64(buf, value)
	return sizeUint64
}

func (uint64Codec) Get(buf []byte) (uint64, []byte) {
	return binary.BigEndian.Uint64(buf), buf[sizeUint64:]
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
	int8Codec  struct{}
	int16Codec struct{}
	int32Codec struct{}
	int64Codec struct{}
)

func (int8Codec) Append(buf []byte, value int8) []byte {
	return append(buf, uint8(math.MinInt8^value))
}

func (int8Codec) Put(buf []byte, value int8) int {
	buf[0] = uint8(math.MinInt8 ^ value)
	return sizeUint8
}

func (int8Codec) Get(buf []byte) (int8, []byte) {
	return math.MinInt8 ^ int8(buf[0]), buf[sizeUint8:]
}

func (int8Codec) RequiresTerminator() bool {
	return false
}

func (int16Codec) Append(buf []byte, value int16) []byte {
	return binary.BigEndian.AppendUint16(buf, uint16(math.MinInt16^value))
}

func (int16Codec) Put(buf []byte, value int16) int {
	binary.BigEndian.PutUint16(buf, uint16(math.MinInt16^value))
	return sizeUint16
}

func (int16Codec) Get(buf []byte) (int16, []byte) {
	return math.MinInt16 ^ int16(binary.BigEndian.Uint16(buf)), buf[sizeUint16:]
}

func (int16Codec) RequiresTerminator() bool {
	return false
}

func (int32Codec) Append(buf []byte, value int32) []byte {
	return binary.BigEndian.AppendUint32(buf, uint32(math.MinInt32^value))
}

func (int32Codec) Put(buf []byte, value int32) int {
	binary.BigEndian.PutUint32(buf, uint32(math.MinInt32^value))
	return sizeUint32
}

func (int32Codec) Get(buf []byte) (int32, []byte) {
	return math.MinInt32 ^ int32(binary.BigEndian.Uint32(buf)), buf[sizeUint32:]
}

func (int32Codec) RequiresTerminator() bool {
	return false
}

func (int64Codec) Append(buf []byte, value int64) []byte {
	return binary.BigEndian.AppendUint64(buf, uint64(math.MinInt64^value))
}

func (int64Codec) Put(buf []byte, value int64) int {
	binary.BigEndian.PutUint64(buf, uint64(math.MinInt64^value))
	return sizeUint64
}

func (int64Codec) Get(buf []byte) (int64, []byte) {
	return math.MinInt64 ^ int64(binary.BigEndian.Uint64(buf)), buf[sizeUint64:]
}

func (int64Codec) RequiresTerminator() bool {
	return false
}
