package internal_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestArrayInt32(t *testing.T) {
	codec := internal.MakeArrayCodec[[5]int32](int32Codec)
	testCodec(t, codec, []testCase[[5]int32]{
		{"[0, 1, -1, min, max]", [5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}, []byte{
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, [5]int32{1, 2, 3, 4, 5})

	codecEmpty := internal.MakeArrayCodec[[0]int32](int32Codec)
	testCodec(t, codecEmpty, []testCase[[0]int32]{
		{"[]", [0]int32{}, []byte(nil)},
	})
}

func TestArrayArrayInt32(t *testing.T) {
	rowCodec := internal.MakeArrayCodec[[5]int32](int32Codec)
	codec := internal.MakeArrayCodec[[3][5]int32](rowCodec)
	testCodec(t, codec, []testCase[[3][5]int32]{
		{"[0, 1, -1, min, max]",
			[3][5]int32{
				{0, 1, -1, math.MinInt32, math.MaxInt32},
				{2, 2, 2, 2, 2},
				{-2, -2, -2, -2, -2},
			},
			[]byte{
				// {0, 1, -1, math.MinInt32, math.MaxInt32}
				0x80, esc, 0x00, esc, 0x00, esc, 0x00,
				0x80, esc, 0x00, esc, 0x00, esc, 0x01,
				0x7F, 0xFF, 0xFF, 0xFF,
				esc, 0x00, esc, 0x00, esc, 0x00, esc, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF,
				term,
				// {2, 2, 2, 2, 2},
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				term,
				// {-2, -2, -2, -2, -2},
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				term,
			}},
	})
	testCodecFail(t, codec, [3][5]int32{
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
	})
}
