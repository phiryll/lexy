package internal

import (
	"encoding/binary"
	"io"
)

// bool and the uints can use the same implementation.
type uintCodec[T bool | uint8 | uint16 | uint32 | uint64] struct{}

func (c uintCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	err := binary.Read(r, binary.BigEndian, &value)
	return value, err
}

func (c uintCodec[T]) Write(value T, w io.Writer) error {
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

/*
type intCodec[T int8 | int16 | int32 | int64] struct {
	mask T
}

func (c intCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	err := binary.Read(r, binary.BigEndian, &value)
	return c.mask ^ value, err
}

func (c intCodec[T]) Write(value T, w io.Writer) error {
	return binary.Write(w, binary.BigEndian, c.mask^value)
}
*/

type Int8Codec struct{}

func (c Int8Codec) Read(r io.Reader) (int8, error) {
	panic("unimplemented")
}

func (c Int8Codec) Write(value int8, w io.Writer) error {
	panic("unimplemented")
}

type Int16Codec struct{}

func (c Int16Codec) Read(r io.Reader) (int16, error) {
	panic("unimplemented")
}

func (c Int16Codec) Write(value int16, w io.Writer) error {
	panic("unimplemented")
}

type Int32Codec struct{}

func (c Int32Codec) Read(r io.Reader) (int32, error) {
	panic("unimplemented")
}

func (c Int32Codec) Write(value int32, w io.Writer) error {
	panic("unimplemented")
}

type Int64Codec struct{}

func (c Int64Codec) Read(r io.Reader) (int64, error) {
	panic("unimplemented")
}

func (c Int64Codec) Write(value int64, w io.Writer) error {
	panic("unimplemented")
}
