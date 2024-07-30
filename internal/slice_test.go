package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func TestSliceInt32(t *testing.T) {
	codec := internal.SliceCodec[[]int32](int32Codec, true)
	testCodec(t, codec, []testCase[[]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []int32{}, []byte{pNonEmpty}},
		{"[0]", []int32{0}, []byte{pNonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", []int32{-1}, []byte{pNonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", []int32{0, 1, -1}, []byte{
			pNonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, []int32{})
}

func TestSliceString(t *testing.T) {
	codec := internal.SliceCodec[[]string](stringCodec, true)
	testCodec(t, codec, []testCase[[]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []string{}, []byte{pNonEmpty}},
		{"[\"\"]", []string{""}, []byte{pNonEmpty, term}},
		{"[a]", []string{"a"}, []byte{pNonEmpty, 'a', term}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			pNonEmpty,
			'a', term,
			term,
			'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []string{})
}

func TestSlicePtrString(t *testing.T) {
	pointerCodec := internal.PointerCodec[*string](stringCodec, true)
	codec := internal.SliceCodec[[]*string](pointerCodec, true)
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []*string{}, []byte{pNonEmpty}},
		{"[nil]", []*string{nil}, []byte{pNonEmpty, pNilFirst, term}},
		{"[*a]", []*string{ptr("a")}, []byte{pNonEmpty, pNonEmpty, 'a', term}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			pNonEmpty,
			pNonEmpty, 'a', term,
			pNilFirst, term,
			pNonEmpty, term,
			pNonEmpty, 'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []*string{})
}

func TestSliceSliceInt32(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]int32](int32Codec, true)
	codec := internal.SliceCodec[[][]int32](sliceCodec, true)
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]int32{}, []byte{pNonEmpty}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{pNonEmpty, pNilFirst, term}},
		{"[[]]", [][]int32{{}}, []byte{pNonEmpty, pNonEmpty, term}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			pNonEmpty,
			// nil
			pNilFirst, term,
			// {0, 1, -1}, escaped
			pNonEmpty,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			term,
			// {}
			pNonEmpty,
			term,
			// {-2, -3}, escaped
			pNonEmpty,
			0x7F, 0xFF, 0xFF, 0xFE,
			0x7F, 0xFF, 0xFF, 0xFD,
			term,
			// nil
			pNilFirst, term,
			// nil
			pNilFirst, term,
		}},
	})
	testCodecFail(t, codec, [][]int32{})
}

func TestSliceSliceString(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]string](stringCodec, true)
	codec := internal.SliceCodec[[][]string](sliceCodec, true)

	testCodec(t, codec, []testCase[[][]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]string{}, []byte{pNonEmpty}},
		{"[nil]", [][]string{[]string(nil)}, []byte{pNonEmpty, pNilFirst, term}},
		{"[[]]", [][]string{{}}, []byte{pNonEmpty, pNonEmpty, term}},
		{"[[\"\"]]", [][]string{{""}}, []byte{
			pNonEmpty, // prefix outer
			pNonEmpty, // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			pNonEmpty,       // outer
			pNilFirst, term, // nil = outer[0]
			pNonEmpty, term, // {} = outer[1]
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			pNonEmpty,       // outer
			pNonEmpty, term, // {} = outer[0]
			pNilFirst, term, // nil = outer[1]
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			pNonEmpty,       // outer
			pNilFirst, term, // nil = outer[0]
			pNonEmpty, // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			pNonEmpty, // outer
			pNonEmpty, // prefix {""}
			esc, term, // "", escaped
			term,            // terminator {""}, within outer
			pNilFirst, term, // nil = outer[1]
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			pNonEmpty,       // outer
			pNonEmpty, term, // {} = outer[0]
			pNonEmpty, // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			pNonEmpty, // outer
			pNonEmpty, // prefix {""}
			esc, term, // "", escaped
			term,            // terminator {""}, within outer
			pNonEmpty, term, // {} = outer[1]
		}},

		// a complex example
		{"[nil, [a, \"\", xyz], nil, [\"\"], []]", [][]string{
			nil,
			{"a", "", "xyz"},
			nil,
			{""},
			{},
		}, []byte{
			pNonEmpty,
			// nil
			pNilFirst, term,
			// {"a", "", "xyz"}, escaped
			pNonEmpty,
			'a', esc, term,
			esc, term,
			'x', 'y', 'z', esc, term,
			term,
			// nil
			pNilFirst, term,
			// {""}, escaped
			pNonEmpty, esc, term,
			term,
			// {}
			pNonEmpty, term,
		}},
	})
}

type sInt []int32

func TestSliceUnderlyingType(t *testing.T) {
	codec := internal.SliceCodec[sInt](int32Codec, true)
	testCodec(t, codec, []testCase[sInt]{
		{"nil", sInt(nil), []byte{pNilFirst}},
		{"empty", sInt([]int32{}), []byte{pNonEmpty}},
		{"[0]", sInt([]int32{0}), []byte{pNonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", sInt([]int32{-1}), []byte{pNonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", sInt([]int32{0, 1, -1}), []byte{
			pNonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, []int32{})
}

func TestSliceNilsLast(t *testing.T) {
	encodeFirst := encoderFor(internal.SliceCodec[[]int32](int32Codec, true))
	encodeLast := encoderFor(internal.SliceCodec[[]int32](int32Codec, false))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst([]int32{-100, 5}),
		encodeFirst([]int32{0}),
		encodeFirst([]int32{0, 0, 0}),
		encodeFirst([]int32{0, 1}),
		encodeFirst([]int32{35}),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast([]int32{-100, 5}),
		encodeLast([]int32{0}),
		encodeLast([]int32{0, 0, 0}),
		encodeLast([]int32{0, 1}),
		encodeLast([]int32{35}),
		encodeLast(nil),
	})
}
