package internal_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
)

// Testing bool, uint/int8, and uint/int32 should be sufficient. 16 and
// 64 bit ints use the same logic.

func TestBool(t *testing.T) {
	codec := internal.UintCodec[bool]{}
	testCodec[bool](t, codec, []testCase[bool]{
		{"false", false, []byte{0}},
		{"true", true, []byte{1}},
	})
	testCodecFail[bool](t, codec, false)
}

func TestUint8(t *testing.T) {
	codec := internal.UintCodec[uint8]{}
	testCodec[uint8](t, codec, []testCase[uint8]{
		{"0x00", 0x00, []byte{0x00}},
		{"0x01", 0x01, []byte{0x01}},
		{"0x7F", 0x7F, []byte{0x7F}},
		{"0x80", 0x80, []byte{0x80}},
		{"0xFF", 0xFF, []byte{0xFF}},
	})
	testCodecFail[uint8](t, codec, 0)
}

func TestUint32(t *testing.T) {
	codec := internal.UintCodec[uint32]{}
	testCodec[uint32](t, codec, []testCase[uint32]{
		{"0x00000000", 0x00000000, []byte{0x00, 0x00, 0x00, 0x00}},
		{"0x00000001", 0x00000001, []byte{0x00, 0x00, 0x00, 0x01}},
		{"0x7FFFFFFF", 0x7FFFFFFF, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0x80000000", 0x80000000, []byte{0x80, 0x00, 0x00, 0x00}},
		{"0xFFFFFFFF", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail[uint32](t, codec, 0)
}

func TestInt8(t *testing.T) {
	codec := internal.IntCodec[int8]{Mask: math.MinInt8}
	testCodec[int8](t, codec, []testCase[int8]{
		{"min", math.MinInt8, []byte{0x00}},
		{"-1", -1, []byte{0x7F}},
		{"0", 0, []byte{0x80}},
		{"+1", 1, []byte{0x81}},
		{"max", math.MaxInt8, []byte{0xFF}},
	})
	testCodecFail[int8](t, codec, 0)
}

func TestInt32(t *testing.T) {
	codec := internal.IntCodec[int32]{Mask: math.MinInt32}
	testCodec[int32](t, codec, []testCase[int32]{
		{"min", math.MinInt32, []byte{0x00, 0x00, 0x00, 0x00}},
		{"-1", -1, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"0", 0, []byte{0x80, 0x00, 0x00, 0x00}},
		{"+1", 1, []byte{0x80, 0x00, 0x00, 0x01}},
		{"max", math.MaxInt32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail[int32](t, codec, 0)
}
