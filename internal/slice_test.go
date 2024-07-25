package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestSliceInt32(t *testing.T) {
	codec := internal.SliceCodec[[]int32](int32Codec)
	testCodec(t, codec, []testCase[[]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []int32{}, []byte{pEmpty}},
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
	codec := internal.SliceCodec[[]string](stringCodec)
	testCodec(t, codec, []testCase[[]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []string{}, []byte{pEmpty}},
		{"[\"\"]", []string{""}, []byte{pNonEmpty, pEmpty, term}},
		{"[a]", []string{"a"}, []byte{pNonEmpty, pNonEmpty, 'a', term}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			pNonEmpty,
			pNonEmpty, 'a', term,
			pEmpty, term,
			pNonEmpty, 'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []string{})
}

func TestSlicePtrString(t *testing.T) {
	pointerCodec := internal.PointerCodec[*string](stringCodec)
	codec := internal.SliceCodec[[]*string](pointerCodec)
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []*string{}, []byte{pEmpty}},
		{"[nil]", []*string{nil}, []byte{pNonEmpty, pNilFirst, term}},
		{"[*a]", []*string{ptr("a")}, []byte{pNonEmpty, pNonEmpty, pNonEmpty, 'a', term}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			pNonEmpty,
			pNonEmpty, pNonEmpty, 'a', term,
			pNilFirst, term,
			pNonEmpty, pEmpty, term,
			pNonEmpty, pNonEmpty, 'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []*string{})
}

func TestSliceSliceInt32(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]int32](int32Codec)
	codec := internal.SliceCodec[[][]int32](sliceCodec)
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]int32{}, []byte{pEmpty}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{pNonEmpty, pNilFirst, term}},
		{"[[]]", [][]int32{{}}, []byte{pNonEmpty, pEmpty, term}},
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
			pEmpty,
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
	sliceCodec := internal.SliceCodec[[]string](stringCodec)
	codec := internal.SliceCodec[[][]string](sliceCodec)

	testCodec(t, codec, []testCase[[][]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]string{}, []byte{pEmpty}},
		{"[nil]", [][]string{[]string(nil)}, []byte{pNonEmpty, pNilFirst, term}},
		{"[[]]", [][]string{{}}, []byte{pNonEmpty, pEmpty, term}},
		{"[[\"\"]]", [][]string{{""}}, []byte{
			pNonEmpty,         // prefix outer
			pNonEmpty,         // prefix {""}
			pEmpty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			pNonEmpty,       // outer
			pNilFirst, term, // nil = outer[0]
			pEmpty, term, // {} = outer[1]
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			pNonEmpty,    // outer
			pEmpty, term, // {} = outer[0]
			pNilFirst, term, // nil = outer[1]
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			pNonEmpty,       // outer
			pNilFirst, term, // nil = outer[0]
			pNonEmpty,         // prefix {""}
			pEmpty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			pNonEmpty,         // outer
			pNonEmpty,         // prefix {""}
			pEmpty, esc, term, // "", escaped
			term,            // terminator {""}, within outer
			pNilFirst, term, // nil = outer[1]
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			pNonEmpty,    // outer
			pEmpty, term, // {} = outer[0]
			pNonEmpty,         // prefix {""}
			pEmpty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			pNonEmpty,         // outer
			pNonEmpty,         // prefix {""}
			pEmpty, esc, term, // "", escaped
			term,         // terminator {""}, within outer
			pEmpty, term, // {} = outer[1]
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
			pNonEmpty, 'a', esc, term,
			pEmpty, esc, term,
			pNonEmpty, 'x', 'y', 'z', esc, term,
			term,
			// nil
			pNilFirst, term,
			// {""}, escaped
			pNonEmpty, pEmpty, esc, term,
			term,
			// {}
			pEmpty, term,
		}},
	})
}

type sInt []int32

func TestSliceUnderlyingType(t *testing.T) {
	codec := internal.SliceCodec[sInt](int32Codec)
	testCodec(t, codec, []testCase[sInt]{
		{"nil", sInt(nil), []byte{pNilFirst}},
		{"empty", sInt([]int32{}), []byte{pEmpty}},
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
