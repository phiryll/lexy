package lexy_test

import (
	"math"
	"testing"
	"time"

	"github.com/phiryll/lexy"
)

// Testing bool, uint/int, uint/int8, and uint/int32 should be sufficient.

func TestBool(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Bool(), []testCase[bool]{
		{"false", false, []byte{0}},
		{"true", true, []byte{1}},
	})
	testCodecFail(t, lexy.Bool(), false)
}

func TestUint8(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Uint8(), []testCase[uint8]{
		{"0x00", 0x00, []byte{0x00}},
		{"0x01", 0x01, []byte{0x01}},
		{"0x7F", 0x7F, []byte{0x7F}},
		{"0x80", 0x80, []byte{0x80}},
		{"0xFF", 0xFF, []byte{0xFF}},
	})
	testCodecFail(t, lexy.Uint8(), 0)
}

func TestUint32(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Uint32(), []testCase[uint32]{
		{"0x00000000", 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
		{"0x00000001", 0x00000001, []byte{0x00, 0x00, 0x00, 0x01}},
		{"0x7FFFFFFF", 0x7FFFFFFF, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0x80000000", 0x80000000, []byte{0x80, 0x00, 0x00, 0x00}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, lexy.Uint32(), 0)
}

func TestUint(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Uint(), []testCase[uint]{
		{"0", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"1", 1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}},
		// can't go bigger, uints might be 32 bits
	})
	testCodecFail(t, lexy.Uint(), 0)
}

func TestInt8(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Int8(), []testCase[int8]{
		{"min", math.MinInt8, []byte{0x00}},
		{"-1", -1, []byte{0x7F}},
		{"0", 0, []byte{0x80}},
		{"+1", 1, []byte{0x81}},
		{"max", math.MaxInt8, []byte{0xFF}},
	})
	testCodecFail(t, lexy.Int8(), 0)
}

func TestInt32(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Int32(), []testCase[int32]{
		{"min", math.MinInt32, []byte{0x00, 0x00, 0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, lexy.Int32(), 0)
}

func TestInt(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Int(), []testCase[int]{
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
	})
	testCodecFail(t, lexy.Int(), 0)
}

func TestDuration(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Duration(), []testCase[time.Duration]{
		{"min", math.MinInt64 * time.Nanosecond, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-1", -time.Nanosecond, []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"0", 0 * time.Nanosecond, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"+1", time.Nanosecond, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt64 * time.Nanosecond, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, lexy.Duration(), 0)
}

type aBool bool

func TestBoolUnderlyingType(t *testing.T) {
	t.Parallel()
	codec := lexy.MakeBool[aBool]()
	testCodec(t, codec, []testCase[aBool]{
		{"false", aBool(false), []byte{0}},
		{"true", aBool(true), []byte{1}},
	})
	testCodecFail(t, codec, false)
}
