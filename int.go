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
	uint8Codec  struct{}
	uint16Codec struct{}
	uint32Codec struct{}
	uint64Codec struct{}
)

const (
	uint8Size  = 1
	uint16Size = 2
	uint32Size = 4
	uint64Size = 8
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

func (boolCodec) Append(buf []byte, value bool) []byte {
	if value {
		return append(buf, 1)
	}
	return append(buf, 0)
}

func (boolCodec) Put(buf []byte, value bool) int {
	if value {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	return uint8Size
}

func (boolCodec) Get(buf []byte) (bool, int) {
	return buf[0] != 0, uint8Size
}

func (boolCodec) MaxSize() int {
	return uint8Size
}

func (boolCodec) RequiresTerminator() bool {
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

func (uint8Codec) Append(buf []byte, value uint8) []byte {
	return append(buf, value)
}

func (uint8Codec) Put(buf []byte, value uint8) int {
	buf[0] = value
	return uint8Size
}

func (uint8Codec) Get(buf []byte) (uint8, int) {
	return buf[0], uint8Size
}

func (uint8Codec) MaxSize() int {
	return uint8Size
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

func (uint16Codec) Append(buf []byte, value uint16) []byte {
	return binary.BigEndian.AppendUint16(buf, value)
}

func (uint16Codec) Put(buf []byte, value uint16) int {
	binary.BigEndian.PutUint16(buf, value)
	return uint16Size
}

func (uint16Codec) Get(buf []byte) (uint16, int) {
	return binary.BigEndian.Uint16(buf), uint16Size
}

func (uint16Codec) MaxSize() int {
	return uint16Size
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

func (uint32Codec) Append(buf []byte, value uint32) []byte {
	return binary.BigEndian.AppendUint32(buf, value)
}

func (uint32Codec) Put(buf []byte, value uint32) int {
	binary.BigEndian.PutUint32(buf, value)
	return uint32Size
}

func (uint32Codec) Get(buf []byte) (uint32, int) {
	return binary.BigEndian.Uint32(buf), uint32Size
}

func (uint32Codec) MaxSize() int {
	return uint32Size
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

func (uint64Codec) Append(buf []byte, value uint64) []byte {
	return binary.BigEndian.AppendUint64(buf, value)
}

func (uint64Codec) Put(buf []byte, value uint64) int {
	binary.BigEndian.PutUint64(buf, value)
	return uint64Size
}

func (uint64Codec) Get(buf []byte) (uint64, int) {
	return binary.BigEndian.Uint64(buf), uint64Size
}

func (uint64Codec) MaxSize() int {
	return uint64Size
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

func (int8Codec) Append(buf []byte, value int8) []byte {
	return append(buf, uint8(math.MinInt8^value))
}

func (int8Codec) Put(buf []byte, value int8) int {
	buf[0] = uint8(math.MinInt8 ^ value)
	return uint8Size
}

func (int8Codec) Get(buf []byte) (int8, int) {
	return math.MinInt8 ^ int8(buf[0]), uint8Size
}

func (int8Codec) MaxSize() int {
	return uint8Size
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

func (int16Codec) Append(buf []byte, value int16) []byte {
	return binary.BigEndian.AppendUint16(buf, uint16(math.MinInt16^value))
}

func (int16Codec) Put(buf []byte, value int16) int {
	binary.BigEndian.PutUint16(buf, uint16(math.MinInt16^value))
	return uint16Size
}

func (int16Codec) Get(buf []byte) (int16, int) {
	return math.MinInt16 ^ int16(binary.BigEndian.Uint16(buf)), uint16Size
}

func (int16Codec) MaxSize() int {
	return uint16Size
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

func (int32Codec) Append(buf []byte, value int32) []byte {
	return binary.BigEndian.AppendUint32(buf, uint32(math.MinInt32^value))
}

func (int32Codec) Put(buf []byte, value int32) int {
	binary.BigEndian.PutUint32(buf, uint32(math.MinInt32^value))
	return uint32Size
}

func (int32Codec) Get(buf []byte) (int32, int) {
	return math.MinInt32 ^ int32(binary.BigEndian.Uint32(buf)), uint32Size
}

func (int32Codec) MaxSize() int {
	return uint32Size
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

func (int64Codec) Append(buf []byte, value int64) []byte {
	return binary.BigEndian.AppendUint64(buf, uint64(math.MinInt64^value))
}

func (int64Codec) Put(buf []byte, value int64) int {
	binary.BigEndian.PutUint64(buf, uint64(math.MinInt64^value))
	return uint64Size
}

func (int64Codec) Get(buf []byte) (int64, int) {
	return math.MinInt64 ^ int64(binary.BigEndian.Uint64(buf)), uint64Size
}

func (int64Codec) MaxSize() int {
	return uint64Size
}

func (int64Codec) RequiresTerminator() bool {
	return false
}
