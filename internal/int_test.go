package internal

import (
	"math"
	"testing"
)

// Testing bool, uint/int8, and uint/int32 should be sufficient. 16 and
// 64 bit ints use the same logic.

func TestBoolCodec(t *testing.T) {
	testCodec[bool](t, UintCodec[bool]{}, []testCase[bool]{
		{"false", false, []byte{0}},
		{"true", true, []byte{1}},
	})
}

func TestUint8Codec(t *testing.T) {
	testCodec[uint8](t, UintCodec[uint8]{}, []testCase[uint8]{
		{"0x00", 0x00, []byte{0x00}},
		{"0x01", 0x01, []byte{0x01}},
		{"0x7F", 0x7F, []byte{0x7F}},
		{"0x80", 0x80, []byte{0x80}},
		{"0xFF", 0xFF, []byte{0xFF}},
	})
}

func TestUint32Codec(t *testing.T) {
	testCodec[uint32](t, UintCodec[uint32]{}, []testCase[uint32]{
		{"0x00000000", 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
		{"0x00000001", 0x00000001, []byte{0x00, 0x00, 0x00, 0x01}},
		{"0x7FFFFFFF", 0x7FFFFFFF, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0x80000000", 0x80000000, []byte{0x80, 0x00, 0x00, 0x00}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
}

func TestInt8Codec(t *testing.T) {
	testCodec[int8](t, IntCodec[int8]{Mask: math.MinInt8}, []testCase[int8]{
		{"min", math.MinInt8, []byte{0x00}},
		{"-1", -1, []byte{0x7F}},
		{"0", 0, []byte{0x80}},
		{"+1", 1, []byte{0x81}},
		{"max", math.MaxInt8, []byte{0xFF}},
	})
}

func TestInt32Codec(t *testing.T) {
	testCodec[int32](t, IntCodec[int32]{Mask: math.MinInt32}, []testCase[int32]{
		{"min", math.MinInt32, []byte{0x00, 0x00, 0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
}
