package internal_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func TestArrayInt32(t *testing.T) {
	codec := internal.ArrayCodec[[5]int32](int32Codec)
	testCodec(t, codec, []testCase[[5]int32]{
		{"[0, 1, -1, min, max]", [5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}, []byte{
			pNonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, [5]int32{1, 2, 3, 4, 5})

	codecEmpty := internal.ArrayCodec[[0]int32](int32Codec)
	testCodec(t, codecEmpty, []testCase[[0]int32]{
		{"[]", [0]int32{}, []byte{pNonEmpty}},
	})
}

func TestArrayArrayInt32(t *testing.T) {
	rowCodec := internal.ArrayCodec[[5]int32](int32Codec)
	codec := internal.ArrayCodec[[3][5]int32](rowCodec)
	testCodec(t, codec, []testCase[[3][5]int32]{
		{"[0, 1, -1, min, max], ...",
			[3][5]int32{
				{0, 1, -1, math.MinInt32, math.MaxInt32},
				{2, 2, 2, 2, 2},
				{-2, -2, -2, -2, -2},
			},
			[]byte{
				pNonEmpty,
				// {0, 1, -1, math.MinInt32, math.MaxInt32}
				pNonEmpty,
				0x80, esc, 0x00, esc, 0x00, esc, 0x00,
				0x80, esc, 0x00, esc, 0x00, esc, 0x01,
				0x7F, 0xFF, 0xFF, 0xFF,
				esc, 0x00, esc, 0x00, esc, 0x00, esc, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF,
				term,
				// {2, 2, 2, 2, 2},
				pNonEmpty,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				0x80, esc, 0x00, esc, 0x00, 0x02,
				term,
				// {-2, -2, -2, -2, -2},
				pNonEmpty,
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
	codec := internal.PointerToArrayCodec[*[5]int32](int32Codec, true)
	testCodec(t, codec, []testCase[*[5]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"[0, 1, -1, min, max]", &[5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}, []byte{
			pNonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, &[5]int32{1, 2, 3, 4, 5})

	codecEmpty := internal.PointerToArrayCodec[*[0]int32](int32Codec, true)
	testCodec(t, codecEmpty, []testCase[*[0]int32]{
		{"[]", &[0]int32{}, []byte{pNonEmpty}},
	})
}

func TestPtrToArrayInt32UnderlyingType(t *testing.T) {
	type aType *[5]int32
	codec := internal.PointerToArrayCodec[aType](int32Codec, true)
	testCodec(t, codec, []testCase[aType]{
		{"nil", nil, []byte{pNilFirst}},
		{"[0, 1, -1, min, max]", aType(&[5]int32{0, 1, -1, math.MinInt32, math.MaxInt32}), []byte{
			pNonEmpty,
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
	rowCodec := internal.PointerToArrayCodec[*[5]int32](int32Codec, true)
	codec := internal.ArrayCodec[[3]*[5]int32](rowCodec)
	testCodec(t, codec, []testCase[[3]*[5]int32]{
		{"[[0, 1, -1, min, max], nil, [-2, -2, -2, -2, -2]]",
			[3]*[5]int32{
				{0, 1, -1, math.MinInt32, math.MaxInt32},
				nil,
				{-2, -2, -2, -2, -2},
			},
			[]byte{
				pNonEmpty,
				// {0, 1, -1, math.MinInt32, math.MaxInt32}
				pNonEmpty,
				0x80, esc, 0x00, esc, 0x00, esc, 0x00,
				0x80, esc, 0x00, esc, 0x00, esc, 0x01,
				0x7F, 0xFF, 0xFF, 0xFF,
				esc, 0x00, esc, 0x00, esc, 0x00, esc, 0x00,
				0xFF, 0xFF, 0xFF, 0xFF,
				term,
				// nil
				pNilFirst,
				term,
				// {-2, -2, -2, -2, -2},
				pNonEmpty,
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

func TestPointerToArrayNilsLast(t *testing.T) {
	encodeFirst := encoderFor(internal.PointerToArrayCodec[*[1]int32](int32Codec, true))
	encodeLast := encoderFor(internal.PointerToArrayCodec[*[1]int32](int32Codec, false))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst(&[1]int32{-100}),
		encodeFirst(&[1]int32{0}),
		encodeFirst(&[1]int32{35}),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast(&[1]int32{-100}),
		encodeLast(&[1]int32{0}),
		encodeLast(&[1]int32{35}),
		encodeLast(nil),
	})
}
