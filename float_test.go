package lexy_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

// Bit masks for the sign, exponent, and matissa of the
// IEEE 754 32- and 64- floating point representations.
const (
	maskSign32 uint32 = 0x80_00_00_00
	maskExp32  uint32 = 0x7F_80_00_00
	maskMant32 uint32 = 0x00_7F_FF_FF
	maskSign64 uint64 = 0x80_00_00_00_00_00_00_00
	maskExp64  uint64 = 0x7F_F0_00_00_00_00_00_00
	maskMant64 uint64 = 0x00_0F_FF_FF_FF_FF_FF_FF
)

func exp32(value float32) uint32 {
	return maskExp32 & math.Float32bits(value)
}

func mant32(value float32) uint32 {
	return maskMant32 & math.Float32bits(value)
}

func exp64(value float64) uint64 {
	return maskExp64 & math.Float64bits(value)
}

func mant64(value float64) uint64 {
	return maskMant64 & math.Float64bits(value)
}

// IEEE 745 32- and 64-bit patterns at the boundaries between different semantics.
// They are defined here in the same order that they should be in when encoded by the Codec.
// "Min" and "Max" in these variable names denote absolute semantic distance from 0.
var (
	negMaxNaN32       = math.Float32frombits(0xFF_FF_FF_FF)
	negMinNaN32       = math.Float32frombits(0xFF_80_00_01)
	negInf32          = math.Float32frombits(0xFF_80_00_00)
	negMaxNormal32    = math.Float32frombits(0xFF_7F_FF_FF)
	negMinNormal32    = math.Float32frombits(0x80_80_00_00)
	negMaxSubnormal32 = math.Float32frombits(0x80_7F_FF_FF)
	negMinSubnormal32 = math.Float32frombits(0x80_00_00_01)
	negZero32         = math.Float32frombits(0x80_00_00_00)
	posZero32         = math.Float32frombits(0x00_00_00_00)
	posMinSubnormal32 = math.Float32frombits(0x00_00_00_01)
	posMaxSubnormal32 = math.Float32frombits(0x00_7F_FF_FF)
	posMinNormal32    = math.Float32frombits(0x00_80_00_00)
	posMaxNormal32    = math.Float32frombits(0x7F_7F_FF_FF)
	posInf32          = math.Float32frombits(0x7F_80_00_00)
	posMinNaN32       = math.Float32frombits(0x7F_80_00_01)
	posMaxNaN32       = math.Float32frombits(0x7F_FF_FF_FF)

	negMaxNaN64       = math.Float64frombits(0xFF_FF_FF_FF_FF_FF_FF_FF)
	negMinNaN64       = math.Float64frombits(0xFF_F0_00_00_00_00_00_01)
	negInf64          = math.Float64frombits(0xFF_F0_00_00_00_00_00_00)
	negMaxNormal64    = math.Float64frombits(0xFF_EF_FF_FF_FF_FF_FF_FF)
	negMinNormal64    = math.Float64frombits(0x80_10_00_00_00_00_00_00)
	negMaxSubnormal64 = math.Float64frombits(0x80_0F_FF_FF_FF_FF_FF_FF)
	negMinSubnormal64 = math.Float64frombits(0x80_00_00_00_00_00_00_01)
	negZero64         = math.Float64frombits(0x80_00_00_00_00_00_00_00)
	posZero64         = math.Float64frombits(0x00_00_00_00_00_00_00_00)
	posMinSubnormal64 = math.Float64frombits(0x00_00_00_00_00_00_00_01)
	posMaxSubnormal64 = math.Float64frombits(0x00_0F_FF_FF_FF_FF_FF_FF)
	posMinNormal64    = math.Float64frombits(0x00_10_00_00_00_00_00_00)
	posMaxNormal64    = math.Float64frombits(0x7F_EF_FF_FF_FF_FF_FF_FF)
	posInf64          = math.Float64frombits(0x7F_F0_00_00_00_00_00_00)
	posMinNaN64       = math.Float64frombits(0x7F_F0_00_00_00_00_00_01)
	posMaxNaN64       = math.Float64frombits(0x7F_FF_FF_FF_FF_FF_FF_FF)
)

// float32 testCases in increasing order.
var float32TestCases = []testCase[float32]{
	{"-max NaN", negMaxNaN32, nil},
	{"-min NaN", negMinNaN32, nil},
	{"-Inf", negInf32, nil},
	{"-max normal", negMaxNormal32, nil},
	{"-min normal", negMinNormal32, nil},
	{"-max subnormal", negMaxSubnormal32, nil},
	{"-min subnormal", negMinSubnormal32, nil},
	{"-0", negZero32, nil},
	{"+0", posZero32, nil},
	{"+min subnormal", posMinSubnormal32, nil},
	{"+max subnormal", posMaxSubnormal32, nil},
	{"+min normal", posMinNormal32, nil},
	{"+max normal", posMaxNormal32, nil},
	{"+Inf", posInf32, nil},
	{"+min NaN", posMinNaN32, nil},
	{"+max NaN", posMaxNaN32, nil},
}

// float32 testCases in increasing order without NaNs.
var float32NumberTestCases = []testCase[float32]{
	{"-Inf", negInf32, nil},
	{"-max normal", negMaxNormal32, nil},
	{"-min normal", negMinNormal32, nil},
	{"-max subnormal", negMaxSubnormal32, nil},
	{"-min subnormal", negMinSubnormal32, nil},
	{"-0", negZero32, nil},
	{"+0", posZero32, nil},
	{"+min subnormal", posMinSubnormal32, nil},
	{"+max subnormal", posMaxSubnormal32, nil},
	{"+min normal", posMinNormal32, nil},
	{"+max normal", posMaxNormal32, nil},
	{"+Inf", posInf32, nil},
}

// float64 testCases in increasing order.
var float64TestCases = []testCase[float64]{
	{"-max NaN", negMaxNaN64, nil},
	{"-min NaN", negMinNaN64, nil},
	{"-Inf", negInf64, nil},
	{"-max normal", negMaxNormal64, nil},
	{"-min normal", negMinNormal64, nil},
	{"-max subnormal", negMaxSubnormal64, nil},
	{"-min subnormal", negMinSubnormal64, nil},
	{"-0", negZero64, nil},
	{"+0", posZero64, nil},
	{"+min subnormal", posMinSubnormal64, nil},
	{"+max subnormal", posMaxSubnormal64, nil},
	{"+min normal", posMinNormal64, nil},
	{"+max normal", posMaxNormal64, nil},
	{"+Inf", posInf64, nil},
	{"+min NaN", posMinNaN64, nil},
	{"+max NaN", posMaxNaN64, nil},
}

// float64 testCases in increasing order, without NaNs.
var float64NumberTestCases = []testCase[float64]{
	{"-Inf", negInf64, nil},
	{"-max normal", negMaxNormal64, nil},
	{"-min normal", negMinNormal64, nil},
	{"-max subnormal", negMaxSubnormal64, nil},
	{"-min subnormal", negMinSubnormal64, nil},
	{"-0", negZero64, nil},
	{"+0", posZero64, nil},
	{"+min subnormal", posMinSubnormal64, nil},
	{"+max subnormal", posMaxSubnormal64, nil},
	{"+min normal", posMinNormal64, nil},
	{"+max normal", posMaxNormal64, nil},
	{"+Inf", posInf64, nil},
}

// Some of these tests are to make sure I didn't fat-finger anything,
// which I absolutely did the first time around.

// Assert that the bits of b are exactly one more than the bits of a.
func assertNext32(t *testing.T, a, b float32) {
	t.Helper()
	assert.Equal(t, math.Float32bits(a)+1, math.Float32bits(b))
}

// Assert that the bits of b are exactly one more than the bits of a.
func assertNext64(t *testing.T, a, b float64) {
	t.Helper()
	assert.Equal(t, math.Float64bits(a)+1, math.Float64bits(b))
}

// Test the expected ordering of the IEEE 754 32-bit encodings as uint32.
// This ensures that none of the ranges defined by the bit patterns overlap.
//
//nolint:dupl
func TestIEEEOrdering32(t *testing.T) {
	t.Parallel()
	assert.IsIncreasing(t, []uint32{
		math.Float32bits(posZero32),
		math.Float32bits(posMinSubnormal32),
		math.Float32bits(posMaxSubnormal32),
		math.Float32bits(posMinNormal32),
		math.Float32bits(posMaxNormal32),
		math.Float32bits(posInf32),
		math.Float32bits(posMinNaN32),
		math.Float32bits(posMaxNaN32),
		math.Float32bits(negZero32),
		math.Float32bits(negMinSubnormal32),
		math.Float32bits(negMaxSubnormal32),
		math.Float32bits(negMinNormal32),
		math.Float32bits(negMaxNormal32),
		math.Float32bits(negInf32),
		math.Float32bits(negMinNaN32),
		math.Float32bits(negMaxNaN32),
	})

	// Verify the above IsIncreasing test covers the entire range of uint32s.
	assert.Equal(t, uint32(0), math.Float32bits(posZero32))
	assert.Equal(t, uint32(math.MaxUint32), math.Float32bits(negMaxNaN32))

	assertNext32(t, posZero32, posMinSubnormal32)
	assertNext32(t, posMaxSubnormal32, posMinNormal32)
	assertNext32(t, posMaxNormal32, posInf32)
	assertNext32(t, posInf32, posMinNaN32)
	assertNext32(t, posMaxNaN32, negZero32)
	assertNext32(t, negZero32, negMinSubnormal32)
	assertNext32(t, negMaxSubnormal32, negMinNormal32)
	assertNext32(t, negMaxNormal32, negInf32)
	assertNext32(t, negInf32, negMinNaN32)
}

// Test semantic ordering for orderable values (not the NaNs).
// This also tests that all the normal/subnormal constants are neither NaN nor infinite,
// because NaNs are not orderable, and negInf32 and posInf32 are at the extremes of this test.
func TestSemanticOrdering32(t *testing.T) {
	t.Parallel()
	assert.IsIncreasing(t, []float32{
		negInf32,
		negMaxNormal32,
		negMinNormal32,
		negMaxSubnormal32,
		negMinSubnormal32,
		posZero32,
		posMinSubnormal32,
		posMaxSubnormal32,
		posMinNormal32,
		posMaxNormal32,
		posInf32,
	})
}

// Test that the bit patterns are what their names say they are.
func TestNames32(t *testing.T) {
	t.Parallel()
	// Testable exact values
	assert.Equal(t, math.Inf(-1), float64(negInf32), "-Inf: %x", negInf32)
	assert.Equal(t, math.Inf(1), float64(posInf32), "+Inf: %x", posInf32)
	assert.Equal(t, -float32(math.MaxFloat32), negMaxNormal32, "max negative float32: %x", negMaxNormal32)
	assert.Equal(t, float32(math.MaxFloat32), posMaxNormal32, "max positive float32: %x", posMaxNormal32)
	assert.Equal(t, -float32(math.SmallestNonzeroFloat32), negMinSubnormal32,
		"min negative float32: %x", negMinSubnormal32)
	assert.Equal(t, float32(math.SmallestNonzeroFloat32), posMinSubnormal32,
		"min positive float32: %x", posMinSubnormal32)
	assert.Equal(t, float32(math.Copysign(0.0, -1.0)), negZero32, "should be -0.0: %x", negZero32)
	assert.Equal(t, float32(math.Copysign(0.0, 1.0)), posZero32, "should be +0.0: %x", posZero32)

	// Test NaNs
	for _, x := range []float32{negMaxNaN32, negMinNaN32, posMinNaN32, posMaxNaN32} {
		assert.True(t, math.IsNaN(float64(x)), "should be NaN: %x", x)
	}

	// Test exponents and matissas
	for _, x := range []float32{negMaxNormal32, negMinNormal32, posMinNormal32, posMaxNormal32} {
		assert.NotEqual(t, uint32(0), exp32(x), "non-zero normal numbers should have a non-zero exponent: %x", x)
		assert.NotEqual(t, maskExp32, exp32(x), "non-zero normal numbers should have a non-0xFF exponent: %x", x)
	}
	for _, x := range []float32{negMaxSubnormal32, negMinSubnormal32, posMinSubnormal32, posMaxSubnormal32} {
		assert.Equal(t, uint32(0), exp32(x), "subnormal numbers should have a zero exponent: %x", x)
		assert.NotEqual(t, uint32(0), mant32(x), "subnormal numbers should have a non-zero mantissa: %x", x)
	}
}

func TestFloat32(t *testing.T) {
	t.Parallel()
	codec := lexy.Float32()
	testCodec(t, codec, fillTestData(codec, float32NumberTestCases))
}

// Test that the encoded forms have the right lexicographical ordering.
func TestFloat32CodecOrdering(t *testing.T) {
	t.Parallel()
	codec := lexy.Float32()
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, codec.Append(nil, negMaxNaN32))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF}, codec.Append(nil, posMaxNaN32))
	testOrdering(t, codec, float32TestCases)
}

// The 64-bit float tests are the same as the 32-bit float tests.

//nolint:dupl
func TestIEEEOrdering64(t *testing.T) {
	t.Parallel()
	assert.IsIncreasing(t, []uint64{
		math.Float64bits(posZero64),
		math.Float64bits(posMinSubnormal64),
		math.Float64bits(posMaxSubnormal64),
		math.Float64bits(posMinNormal64),
		math.Float64bits(posMaxNormal64),
		math.Float64bits(posInf64),
		math.Float64bits(posMinNaN64),
		math.Float64bits(posMaxNaN64),
		math.Float64bits(negZero64),
		math.Float64bits(negMinSubnormal64),
		math.Float64bits(negMaxSubnormal64),
		math.Float64bits(negMinNormal64),
		math.Float64bits(negMaxNormal64),
		math.Float64bits(negInf64),
		math.Float64bits(negMinNaN64),
		math.Float64bits(negMaxNaN64),
	})

	assert.Equal(t, uint64(0), math.Float64bits(posZero64))
	assert.Equal(t, uint64(math.MaxUint64), math.Float64bits(negMaxNaN64))

	assertNext64(t, posZero64, posMinSubnormal64)
	assertNext64(t, posMaxSubnormal64, posMinNormal64)
	assertNext64(t, posMaxNormal64, posInf64)
	assertNext64(t, posInf64, posMinNaN64)
	assertNext64(t, posMaxNaN64, negZero64)
	assertNext64(t, negZero64, negMinSubnormal64)
	assertNext64(t, negMaxSubnormal64, negMinNormal64)
	assertNext64(t, negMaxNormal64, negInf64)
	assertNext64(t, negInf64, negMinNaN64)
}

func TestSemanticOrdering64(t *testing.T) {
	t.Parallel()
	assert.IsIncreasing(t, []float64{
		negInf64,
		negMaxNormal64,
		negMinNormal64,
		negMaxSubnormal64,
		negMinSubnormal64,
		posZero64,
		posMinSubnormal64,
		posMaxSubnormal64,
		posMinNormal64,
		posMaxNormal64,
		posInf64,
	})
}

func TestNames64(t *testing.T) {
	t.Parallel()
	assert.Equal(t, math.Inf(-1), negInf64, "-Inf: %x", negInf64)
	assert.Equal(t, math.Inf(1), posInf64, "+Inf: %x", posInf64)
	assert.Equal(t, -math.MaxFloat64, negMaxNormal64, "max negative float64: %x", negMaxNormal64)
	assert.Equal(t, math.MaxFloat64, posMaxNormal64, "max positive float64: %x", posMaxNormal64)
	assert.Equal(t, -math.SmallestNonzeroFloat64, negMinSubnormal64, "min negative float64: %x", negMinSubnormal64)
	assert.Equal(t, math.SmallestNonzeroFloat64, posMinSubnormal64, "min positive float64: %x", posMinSubnormal64)
	assert.Equal(t, math.Copysign(0.0, -1.0), negZero64, "should be -0.0: %x", negZero64)
	assert.Equal(t, math.Copysign(0.0, 1.0), posZero64, "should be +0.0: %x", posZero64)

	for _, x := range []float64{negMaxNaN64, negMinNaN64, posMinNaN64, posMaxNaN64} {
		assert.True(t, math.IsNaN(x), "should be NaN: %x", x)
	}
	for _, x := range []float64{negMaxNormal64, negMinNormal64, posMinNormal64, posMaxNormal64} {
		assert.NotEqual(t, uint64(0), exp64(x), "non-zero normal numbers should have a non-zero exponent: %x", x)
		assert.NotEqual(t, maskExp64, exp64(x), "non-zero normal numbers should have a non-0xFF exponent: %x", x)
	}
	for _, x := range []float64{negMaxSubnormal64, negMinSubnormal64, posMinSubnormal64, posMaxSubnormal64} {
		assert.Equal(t, uint64(0), exp64(x), "subnormal numbers should have a zero exponent: %x", x)
		assert.NotEqual(t, uint64(0), mant64(x), "subnormal numbers should have a non-zero mantissa: %x", x)
	}
}

func TestFloat64(t *testing.T) {
	t.Parallel()
	codec := lexy.Float64()
	testCodec(t, codec, fillTestData(codec, float64NumberTestCases))
}

func TestFloat64CodecOrdering(t *testing.T) {
	t.Parallel()
	codec := lexy.Float64()
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, codec.Append(nil, negMaxNaN64))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, codec.Append(nil, posMaxNaN64))
	testOrdering(t, codec, float64TestCases)
}
