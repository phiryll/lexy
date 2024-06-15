package internal

import (
	"encoding/binary"
	"io"
	"math"
)

// This Codec encodes float32s to be consistent with the following
// ordering. Since the standard does not define their binary formats, no
// distinction is made between quiet and signaling NaNs.
//
//	-NaN
//	-Infinity
//	-x, for normal negative numbers x
//	-s, for subnormal negative numbers s
//	-0
//	+0
//	+s, for subnormal positive numbers s
//	+x, for normal positive numbers x
//	+Infinity
//	+NaN
//
// The rest of this comment contains details about IEEE 754 and how this
// encoding works. Feel free to skip it!
//
// The standard binary format of floats is defined by IEEE 754. The
// represented value is:
//
//	+/-1 * mantissa * 2^exponent
//
// where the binary format is:
//
//	sign - 1 bit
//	    0 := positive, 1 := negative
//	exponent - 8 bits
//	    0x00 :=
//	        +/-0 if the matissa is zero
//	        a subnormal number if the mantissa is not zero, the leading
//	        bit of the mantissa is 0 and is not stored
//	    0x01-0xFE := normal exponent = value - 127 (-126 to +127)
//	        the mantissa encodes a value in the range [1.0, 2.0), with
//	        the leading 1 bit being implicit and not stored
//	    0xFF :=
//	        +/- infinity if the mantissa is zero
//	        NaN if the mantissa is not zero
//	mantissa - 23 bits
//	    see above for interpretation
//
// Because of the special values, IEEE 754 defines ordering in a way
// that is not consistent with Codec's semantics. In particular, -0 and
// +0 are equal and NaN is not comparable to anything, even another NaN.
// These will all by encoded to some []byte by this Codec, and will be
// comparable in that representation.
//
// By design, a float32's bits interpreted as a signed-magnitude int
// (not the normal 2's complement int) will result in the right
// ordering. To convert a float32's bits to the correct []byte
// lexicographical ordering, we need to:
//
//	flip the high bit if the sign bit is 0
//	flip all the bits if the sign bit is 1
type Float32Codec struct{}

const highBit32 = 0x80_00_00_00
const allBits32 = 0xFF_FF_FF_FF

func (c Float32Codec) Read(r io.Reader) (float32, error) {
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

func (c Float32Codec) Write(w io.Writer, value float32) error {
	bits := math.Float32bits(value)
	if bits&highBit32 == 0 {
		bits ^= highBit32
	} else {
		bits ^= allBits32
	}
	return binary.Write(w, binary.BigEndian, bits)
}

// This has the same general behavior as Float32Codec. The IEEE 754
// format differs slightly, but is otherwise analagous.
//
//	sign - 1 bit
//	exponent - 11 bits
//	mantissa - 52 bits
type Float64Codec struct{}

const highBit64 = 0x80_00_00_00_00_00_00_00
const allBits64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

func (c Float64Codec) Read(r io.Reader) (float64, error) {
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

func (c Float64Codec) Write(w io.Writer, value float64) error {
	bits := math.Float64bits(value)
	if bits&highBit64 == 0 {
		bits ^= highBit64
	} else {
		bits ^= allBits64
	}
	return binary.Write(w, binary.BigEndian, bits)
}
