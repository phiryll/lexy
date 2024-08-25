package lexy

import (
	"math"
)

// Codecs for float32 and float64 types.
//
// No distinction is made between quiet and signaling NaNs.
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
// The rest of this comment contains details about IEEE 754 and how this encoding works.
// Feel free to skip it!
//
// IEEE 754 defines the represented floating point value as:
//
//	+/-1 * mantissa * 2^exponent
//
// where the binary format for 32 bit floats, from high bit to low, is:
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
// The IEEE 754 format for float64 differs slightly, but is otherwise analogous.
//
//	sign - 1 bit
//	exponent - 11 bits
//	mantissa - 52 bits
//
// IEEE 754 defines ordering in a way that is inconsistent with Codec's semantics:
//   - -0.0 and +0.0 are equal
//   - NaN is not comparable to anything, even another NaN
//   - There are many bit patterns for NaN
//
// These will all by encoded by these Codecs, and will be comparable in that representation.
// Every NaN bit pattern will be encoded differently, and will therefore be unequal and comparable.
//
// By design, a float's bits interpreted as a signed-magnitude int
// (not the normal 2's complement int) will result in the right ordering.
// To give the correct unsigned binary lexicographical ordering, we need to:
//
//	flip the high bit if the sign bit is 0
//	flip all the bits if the sign bit is 1
type (
	float32Codec struct{}
	float64Codec struct{}
)

const (
	highBit32 uint32 = 0x80_00_00_00
	allBits32 uint32 = 0xFF_FF_FF_FF
	highBit64 uint64 = 0x80_00_00_00_00_00_00_00
	allBits64 uint64 = 0xFF_FF_FF_FF_FF_FF_FF_FF
)

func float32ToBits(value float32) uint32 {
	bits := math.Float32bits(value)
	if bits&highBit32 == 0 {
		return bits ^ highBit32
	}
	return bits ^ allBits32
}

func float32FromBits(bits uint32) float32 {
	if bits&highBit32 == 0 {
		return math.Float32frombits(bits ^ allBits32)
	}
	return math.Float32frombits(bits ^ highBit32)
}

func float64ToBits(value float64) uint64 {
	bits := math.Float64bits(value)
	if bits&highBit64 == 0 {
		return bits ^ highBit64
	}
	return bits ^ allBits64
}

func float64FromBits(bits uint64) float64 {
	if bits&highBit64 == 0 {
		return math.Float64frombits(bits ^ allBits64)
	}
	return math.Float64frombits(bits ^ highBit64)
}

func (float32Codec) Append(buf []byte, value float32) []byte {
	return stdUint32.Append(buf, float32ToBits(value))
}

func (float32Codec) Put(buf []byte, value float32) []byte {
	return stdUint32.Put(buf, float32ToBits(value))
}

func (float32Codec) Get(buf []byte) (float32, []byte) {
	bits, buf := stdUint32.Get(buf)
	return float32FromBits(bits), buf
}

func (float32Codec) RequiresTerminator() bool {
	return false
}

func (float64Codec) Append(buf []byte, value float64) []byte {
	return stdUint64.Append(buf, float64ToBits(value))
}

func (float64Codec) Put(buf []byte, value float64) []byte {
	return stdUint64.Put(buf, float64ToBits(value))
}

func (float64Codec) Get(buf []byte) (float64, []byte) {
	bits, buf := stdUint64.Get(buf)
	return float64FromBits(bits), buf
}

func (float64Codec) RequiresTerminator() bool {
	return false
}
