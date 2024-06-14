package internal

import (
	"encoding/binary"
	"io"
)

// Unsigned ints and bool can use the same implementation. This encoding
// is merely the value in big-endian order. Instances are instantiated
// as:
//
//	UintCodec[<type>]{}
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

// Signed ints use nearly the same implementation. This encoding is the
// value with its sign bit flipped in big-endian order. They should map:
//
//	0x8000... -> 0x0000...  most negative
//	0xFFFF... -> 0x7FFF...  -1
//	0x0000... -> 0x8000...  0
//	0x0000..1 -> 0x8000..1  1
//	0x7FFF... -> 0xFFFF...  most positive
//
// Mask should be the minimum signed value for that type, because that
// happens to be of the form 0x8000... for fixed length signed integral
// types. Instances are instantiated as:
//
//	IntCodec[<type>]{Mask: math.Min<type>}
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
