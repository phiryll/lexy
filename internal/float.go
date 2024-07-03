package internal

import (
	"encoding/binary"
	"io"
	"math"
)

var (
	Float32Codec codec[float32] = float32Codec{}
	Float64Codec codec[float64] = float64Codec{}
)

// float32Codec is the Codec for float32.
//
// The order of the encoded values is:
//
//	-NaN
//	-Infinity
//	-x, for normal negative numbers x
//	-s, for subnormal negative numbers s
//	-0.0
//	+0.0
//	+s, for subnormal positive numbers s
//	+x, for normal positive numbers x
//	+Infinity
//	+NaN
//
// No distinction is made between quiet and signaling NaNs.
//
// The rest of this comment contains details about IEEE 754 and how this encoding works.
// Feel free to skip it!
//
// IEEE 754 defines the represented floating point value as:
//
//	+/-1 * mantissa * 2^exponent
//
// where the binary format, from high bit to low, is:
//
//	sign - 1 bit
//	    0 := positive, 1 := negative
//	exponent - 8 bits
//	    0x00 :=
//	        +/-0 if the matissa is zero
//	        a subnormal number if the mantissa is not zero, the leading bit of the mantissa is implicitly 0
//	    0x01-0xFE := normal exponent = value - 127 (-126 to +127)
//	        the mantissa encodes a value in the range [1.0, 2.0), the leading bit of the mantissa is implicitly 1
//	    0xFF :=
//	        +/- infinity if the mantissa is zero
//	        NaN if the mantissa is not zero
//	mantissa - 23 bits
//	    see above for interpretation
//
// IEEE 754 defines ordering in a way that is inconsistent with Codec's semantics:
//   - -0.0 and +0.0 are equal
//   - NaN is not comparable to anything, even another NaN
//   - There are many bit patterns for NaN
//
// These will all by encoded by this Codec, and will be comparable in that representation.
// Every NaN bit pattern will be encoded differently, and will therefore be unequal and comparable.
//
// By design, a float32's bits interpreted as a signed-magnitude int
// (not the normal 2's complement int) will result in the right ordering.
// To give the correct unsigned binary lexicographical ordering, we need to:
//
//	flip the high bit if the sign bit is 0
//	flip all the bits if the sign bit is 1
type float32Codec struct{}

const highBit32 uint32 = 0x80_00_00_00
const allBits32 uint32 = 0xFF_FF_FF_FF

func (c float32Codec) Read(r io.Reader) (float32, error) {
	var bits uint32
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0.0, err
	}
	if bits&highBit32 == 0 {
		bits ^= allBits32
	} else {
		bits ^= highBit32
	}
	return math.Float32frombits(bits), nil
}

func (c float32Codec) Write(w io.Writer, value float32) error {
	bits := math.Float32bits(value)
	if bits&highBit32 == 0 {
		bits ^= highBit32
	} else {
		bits ^= allBits32
	}
	return binary.Write(w, binary.BigEndian, bits)
}

// float64Codec is the Codec for float64, and has the same general behavior as Float32Codec.
//
// The IEEE 754 format differs slightly, but is otherwise analagous.
//
//	sign - 1 bit
//	exponent - 11 bits
//	mantissa - 52 bits
type float64Codec struct{}

const highBit64 uint64 = 0x80_00_00_00_00_00_00_00
const allBits64 uint64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

func (c float64Codec) Read(r io.Reader) (float64, error) {
	var bits uint64
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0.0, err
	}
	if bits&highBit64 == 0 {
		bits ^= allBits64
	} else {
		bits ^= highBit64
	}
	return math.Float64frombits(bits), nil
}

func (c float64Codec) Write(w io.Writer, value float64) error {
	bits := math.Float64bits(value)
	if bits&highBit64 == 0 {
		bits ^= highBit64
	} else {
		bits ^= allBits64
	}
	return binary.Write(w, binary.BigEndian, bits)
}
