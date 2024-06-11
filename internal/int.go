package internal

import (
	"encoding/binary"
	"io"
	"math"
)

// bool and the uints can use the same implementation.
type uintCodec[T bool | uint8 | uint16 | uint32 | uint64] struct{}

func (c uintCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	err := binary.Read(r, binary.BigEndian, &value)
	return value, err
}

func (c uintCodec[T]) Write(w io.Writer, value T) error {
	return binary.Write(w, binary.BigEndian, value)
}

type BoolCodec struct {
	uintCodec[bool]
}

type Uint8Codec struct {
	uintCodec[uint8]
}

type Uint16Codec struct {
	uintCodec[uint16]
}

type Uint32Codec struct {
	uintCodec[uint32]
}

type Uint64Codec struct {
	uintCodec[uint64]
}

// ints can use nearly the same implementation. This should map:
//
//  0x8000... -> 0x0000...  most negative
//  0xFFFF... -> 0x7FFF...  -1
//  0x0000... -> 0x8000...  0
//  0x0000..1 -> 0x8000..1  1
//  0x7FFF... -> 0xFFFF...  most positive
//
// This is merely flipping the high bit. It would be nice to make this a
// little generic like the uints, but there isn't a simple way because
// it would involve type params for (for example) int8 and uint8 (the
// bit mask) and a type cast between the two. It's simpler to just write
// out the code for each case.

type Int8Codec struct{}

func (c Int8Codec) Read(r io.Reader) (int8, error) {
	value, err := Uint8Codec{}.Read(r)
	return math.MinInt8 ^ int8(value), err
}

func (c Int8Codec) Write(w io.Writer, value int8) error {
	return Uint8Codec{}.Write(w, uint8(math.MinInt8^value))
}

type Int16Codec struct{}

func (c Int16Codec) Read(r io.Reader) (int16, error) {
	value, err := Uint16Codec{}.Read(r)
	return math.MinInt16 ^ int16(value), err
}

func (c Int16Codec) Write(w io.Writer, value int16) error {
	return Uint16Codec{}.Write(w, uint16(math.MinInt16^value))
}

type Int32Codec struct{}

func (c Int32Codec) Read(r io.Reader) (int32, error) {
	value, err := Uint32Codec{}.Read(r)
	return math.MinInt32 ^ int32(value), err
}

func (c Int32Codec) Write(w io.Writer, value int32) error {
	return Uint32Codec{}.Write(w, uint32(math.MinInt32^value))
}

type Int64Codec struct{}

func (c Int64Codec) Read(r io.Reader) (int64, error) {
	value, err := Uint64Codec{}.Read(r)
	return math.MinInt64 ^ int64(value), err
}

func (c Int64Codec) Write(w io.Writer, value int64) error {
	return Uint64Codec{}.Write(w, uint64(math.MinInt8^value))
}
