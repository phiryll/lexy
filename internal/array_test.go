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
			nonEmpty,
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
		{"[]", [0]int32{}, []byte{nonEmpty}},
	})
}

func TestArrayArrayInt32(t *testing.T) {
	rowCodec := internal.MakeArrayCodec[[5]int32](int32Codec)
	codec := internal.MakeArrayCodec[[3][5]int32](rowCodec)
	testCodec(t, codec, []testCase[[3][5]int32]{
		{"[0, 1, -1, min, max], ...",
			[3][5]int32{
				{0, 1, -1, math.MinInt32, math.MaxInt32},
				{2, 2, 2, 2, 2},
				{-2, -2, -2, -2, -2},
			},
			[]byte{
				nonEmpty,
				// {0, 1, -1, math.MinInt32, math.MaxInt32}
				nonEmpty,
				0x80, esc, 0x00, esc, 0x00, esc, 0x00,
				0x80, esc, 0x00, esc, 0x00, esc, 0x01,
				0x7F, 0xFF, 0xFF, 0xFF,
				esc, 0x00, esc, 0x00, esc, 0x00, esc, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF,
				term,
				// {2, 2, 2, 2, 2},
				nonEmpty,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				term,
				// {-2, -2, -2, -2, -2},
				nonEmpty,
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

func TestPtrToArrayInt32(t *testing.T) {
	codec := internal.MakePointerToArrayCodec[*[5]int32](int32Codec)
	testCodec(t, codec, []testCase[*[5]int32]{
		{"nil", nil, []byte{pNil}},
		{"[0, 1, -1, min, max]", &[5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}, []byte{
			nonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, &[5]int32{1, 2, 3, 4, 5})

	codecEmpty := internal.MakePointerToArrayCodec[*[0]int32](int32Codec)
	testCodec(t, codecEmpty, []testCase[*[0]int32]{
		{"[]", &[0]int32{}, []byte{nonEmpty}},
	})
}

func TestPtrToArrayInt32UnderlyingType(t *testing.T) {
	type aType *[5]int32
	codec := internal.MakePointerToArrayCodec[aType](int32Codec)
	testCodec(t, codec, []testCase[aType]{
		{"nil", nil, []byte{pNil}},
		{"[0, 1, -1, min, max]", aType(&[5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}), []byte{
			nonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, aType(&[5]int32{1, 2, 3, 4, 5}))
}

func TestArrayPtrToArrayInt32(t *testing.T) {
	rowCodec := internal.MakePointerToArrayCodec[*[5]int32](int32Codec)
	codec := internal.MakeArrayCodec[[3]*[5]int32](rowCodec)
	testCodec(t, codec, []testCase[[3]*[5]int32]{
		{"[[0, 1, -1, min, max], nil, [-2, -2, -2, -2, -2]]",
			[3]*[5]int32{
				{0, 1, -1, math.MinInt32, math.MaxInt32},
				nil,
				{-2, -2, -2, -2, -2},
			},
			[]byte{
				nonEmpty,
				// {0, 1, -1, math.MinInt32, math.MaxInt32}
				nonEmpty,
				0x80, esc, 0x00, esc, 0x00, esc, 0x00,
				0x80, esc, 0x00, esc, 0x00, esc, 0x01,
				0x7F, 0xFF, 0xFF, 0xFF,
				esc, 0x00, esc, 0x00, esc, 0x00, esc, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF,
				term,
				// nil
				pNil,
				term,
				// {-2, -2, -2, -2, -2},
				nonEmpty,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				0x7F, 0xFF, 0xFF, 0xFE,
				term,
			}},
	})
	testCodecFail(t, codec, [3]*[5]int32{
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5},
	})
}
