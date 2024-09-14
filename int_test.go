package lexy_test

import (
	"math"
	"testing"
	"time"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Parallel()
	testBool(t, lexy.Bool())
}

func TestCastBool(t *testing.T) {
	t.Parallel()
	type myBool bool
	testBool(t, lexy.CastBool[myBool]())
}

func TestUint8(t *testing.T) {
	t.Parallel()
	testUint8(t, lexy.Uint8())
}

func TestCastUint8(t *testing.T) {
	t.Parallel()
	type myUint8 uint8
	testUint8(t, lexy.CastUint8[myUint8]())
}

func TestUint16(t *testing.T) {
	t.Parallel()
	testUint16(t, lexy.Uint16())
}

func TestCastUint16(t *testing.T) {
	t.Parallel()
	type myUint16 uint16
	testUint16(t, lexy.CastUint16[myUint16]())
}

func TestUint32(t *testing.T) {
	t.Parallel()
	testUint32(t, lexy.Uint32())
}

func TestCastUint32(t *testing.T) {
	t.Parallel()
	type myUint32 uint32
	testUint32(t, lexy.CastUint32[myUint32]())
}

func TestUint64(t *testing.T) {
	t.Parallel()
	testUint64(t, lexy.Uint64())
}

func TestCastUint64(t *testing.T) {
	t.Parallel()
	type myUint64 uint64
	testUint64(t, lexy.CastUint64[myUint64]())
}

func TestUint(t *testing.T) {
	t.Parallel()
	testUint(t, lexy.Uint())
}

func TestCastUint(t *testing.T) {
	t.Parallel()
	type myUint uint
	testUint(t, lexy.CastUint[myUint]())
}

func TestInt8(t *testing.T) {
	t.Parallel()
	testInt8(t, lexy.Int8())
}

func TestCastInt8(t *testing.T) {
	t.Parallel()
	type myInt8 int8
	testInt8(t, lexy.CastInt8[myInt8]())
}

func TestInt16(t *testing.T) {
	t.Parallel()
	testInt16(t, lexy.Int16())
}

func TestCastInt16(t *testing.T) {
	t.Parallel()
	type myInt16 int16
	testInt16(t, lexy.CastInt16[myInt16]())
}

func TestInt32(t *testing.T) {
	t.Parallel()
	testInt32(t, lexy.Int32())
}

func TestCastInt32(t *testing.T) {
	t.Parallel()
	type myInt32 int32
	testInt32(t, lexy.CastInt32[myInt32]())
}

func TestInt64(t *testing.T) {
	t.Parallel()
	testInt64(t, lexy.Int64())
}

func TestCastInt64(t *testing.T) {
	t.Parallel()
	type myInt64 int64
	testInt64(t, lexy.CastInt64[myInt64]())
}

func TestInt(t *testing.T) {
	t.Parallel()
	testInt(t, lexy.Int())
}

func TestCastInt(t *testing.T) {
	t.Parallel()
	type myInt int
	testInt(t, lexy.CastInt[myInt]())
}

func TestDuration(t *testing.T) {
	t.Parallel()
	codec := lexy.Duration()
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[time.Duration]{
		{"min", math.MinInt64 * time.Nanosecond, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-1", -time.Nanosecond, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0", 0 * time.Nanosecond, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"+1", time.Nanosecond, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt64 * time.Nanosecond, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func testBool[T ~bool](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"false", false, []byte{0}},
		{"true", true, []byte{1}},
	})
}

func testUint8[T ~uint8](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"0x00", 0x00, []byte{0x00}},
		{"0x01", 0x01, []byte{0x01}},
		{"0x7F", 0x7F, []byte{0x7F}},
		{"0x80", 0x80, []byte{0x80}},
		{"0xFF", 0xFF, []byte{0xFF}},
	})
}

func testUint16[T ~uint16](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"0x0000", 0x0000, []byte{0x00, 0x00}},
		{"0x0001", 0x0001, []byte{0x00, 0x01}},
		{"0x7FFF", 0x7FFF, []byte{0x7F, 0xFF}},
		{"0x8000", 0x8000, []byte{0x80, 0x00}},
		{"0xFFFF", 0xFFFF, []byte{0xFF, 0xFF}},
	})
}

func testUint32[T ~uint32](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"0x00000000", 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
		{"0x00000001", 0x00000001, []byte{0x00, 0x00, 0x00, 0x01}},
		{"0x7FFFFFFF", 0x7FFFFFFF, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0x80000000", 0x80000000, []byte{0x80, 0x00, 0x00, 0x00}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func testUint64[T ~uint64](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	// Workaround for the inability to have uint64 constants greater than 0x7FF...
	var half T = 1
	half <<= 63
	var maxValue T
	maxValue = ^maxValue
	testCodec(t, codec, []testCase[T]{
		{"0x0000000000000000", 0x0000000000000000, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"0x0000000000000001", 0x0000000000000001, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"0x7FFFFFFFFFFFFFFF", 0x7FFFFFFFFFFFFFFF, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0x8000000000000000", half, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"0xFFFFFFFFFFFFFFFF", maxValue, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func testUint[T ~uint](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"0", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"1", 1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}},
		// can't go bigger, uints might be 32 bits
	})
}

func testInt8[T ~int8](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"min", math.MinInt8, []byte{0x00}},
		{"-1", -1, []byte{0x7F}},
		{"0", 0, []byte{0x80}},
		{"+1", 1, []byte{0x81}},
		{"max", math.MaxInt8, []byte{0xFF}},
	})
}

func testInt16[T ~int16](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"min", math.MinInt16, []byte{0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF}},
		{"0", 0, []byte{0x80, 0x00}},
		{"+1", 1, []byte{0x80, 0x01}},
		{"max", math.MaxInt16, []byte{0xFF, 0xFF}},
	})
}

func testInt32[T ~int32](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"min", math.MinInt32, []byte{0x00, 0x00, 0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func testInt64[T ~int64](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"min", math.MinInt64, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt64, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func testInt[T ~int](t *testing.T, codec lexy.Codec[T]) {
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[T]{
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
	})
}
