package internal

import (
	"encoding/binary"
	"io"
)

// UintCodec is the Codec for bool and fixed-length unsigned integral types.
//
// These are:
//   - bool
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//
// Instances are instantiated as
//
//	UintCodec[<type>]{}
//
// This encodes a value in big-endian order.
type UintCodec[T bool | uint8 | uint16 | uint32 | uint64] struct{}

func (c UintCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		var zero T
		return zero, err
	}
	return value, nil
}

func (c UintCodec[T]) Write(w io.Writer, value T) error {
	return binary.Write(w, binary.BigEndian, value)
}

// IntCodec is the Codec for fixed-length signed integral types.
//
// These are:
//   - int8
//   - int16
//   - int32
//   - int64
//
// Instances are instantiated as
//
//	IntCodec[<type>]{Mask: math.Min<type>}
//
// This encodes a value by flipping the sign bit and writing in big-endian order.
// You can see that this works from the following signed int -> encoded table.
//
//	0x8000... -> 0x0000...  most negative
//	0xFFFF... -> 0x7FFF...  -1
//	0x0000... -> 0x8000...  0
//	0x0000..1 -> 0x8000..1  1
//	0x7FFF... -> 0xFFFF...  most positive
type IntCodec[T int8 | int16 | int32 | int64] struct {
	Mask T
}

func (c IntCodec[T]) Read(r io.Reader) (T, error) {
	var value T
	if err := binary.Read(r, binary.BigEndian, &value); err != nil {
		var zero T
		return zero, err
	}
	return c.Mask ^ value, nil
}

func (c IntCodec[T]) Write(w io.Writer, value T) error {
	return binary.Write(w, binary.BigEndian, c.Mask^value)
}
