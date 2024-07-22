package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestSliceInt32(t *testing.T) {
	codec := internal.SliceCodec[[]int32](int32Codec)
	testCodec(t, codec, []testCase[[]int32]{
		{"nil", nil, []byte{pNil}},
		{"empty", []int32{}, []byte{empty}},
		{"[0]", []int32{0}, []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", []int32{-1}, []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", []int32{0, 1, -1}, []byte{
			nonEmpty,
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
		{"nil", nil, []byte{pNil}},
		{"empty", []string{}, []byte{empty}},
		{"[\"\"]", []string{""}, []byte{nonEmpty, empty, term}},
		{"[a]", []string{"a"}, []byte{nonEmpty, nonEmpty, 'a', term}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			nonEmpty,
			nonEmpty, 'a', term,
			empty, term,
			nonEmpty, 'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []string{})
}

func TestSlicePtrString(t *testing.T) {
	pointerCodec := internal.PointerCodec[*string](stringCodec)
	codec := internal.SliceCodec[[]*string](pointerCodec)
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{pNil}},
		{"empty", []*string{}, []byte{empty}},
		{"[nil]", []*string{nil}, []byte{nonEmpty, pNil, term}},
		{"[*a]", []*string{ptr("a")}, []byte{nonEmpty, nonEmpty, nonEmpty, 'a', term}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			nonEmpty,
			nonEmpty, nonEmpty, 'a', term,
			pNil, term,
			nonEmpty, empty, term,
			nonEmpty, nonEmpty, 'x', 'y', 'z', term,
		}},
	})
	testCodecFail(t, codec, []*string{})
}

func TestSliceSliceInt32(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]int32](int32Codec)
	codec := internal.SliceCodec[[][]int32](sliceCodec)
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte{pNil}},
		{"[]", [][]int32{}, []byte{empty}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{nonEmpty, pNil, term}},
		{"[[]]", [][]int32{{}}, []byte{nonEmpty, empty, term}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			nonEmpty,
			// nil
			pNil, term,
			// {0, 1, -1}, escaped
			nonEmpty,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			term,
			// {}
			empty,
			term,
			// {-2, -3}, escaped
			nonEmpty,
			0x7F, 0xFF, 0xFF, 0xFE,
			0x7F, 0xFF, 0xFF, 0xFD,
			term,
			// nil
			pNil, term,
			// nil
			pNil, term,
		}},
	})
	testCodecFail(t, codec, [][]int32{})
}

func TestSliceSliceString(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]string](stringCodec)
	codec := internal.SliceCodec[[][]string](sliceCodec)

	testCodec(t, codec, []testCase[[][]string]{
		{"nil", nil, []byte{pNil}},
		{"[]", [][]string{}, []byte{empty}},
		{"[nil]", [][]string{[]string(nil)}, []byte{nonEmpty, pNil, term}},
		{"[[]]", [][]string{{}}, []byte{nonEmpty, empty, term}},
		{"[[\"\"]]", [][]string{{""}}, []byte{
			nonEmpty,         // prefix outer
			nonEmpty,         // prefix {""}
			empty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			nonEmpty,   // outer
			pNil, term, // nil = outer[0]
			empty, term, // {} = outer[1]
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			nonEmpty,    // outer
			empty, term, // {} = outer[0]
			pNil, term, // nil = outer[1]
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			nonEmpty,   // outer
			pNil, term, // nil = outer[0]
			nonEmpty,         // prefix {""}
			empty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			nonEmpty,         // outer
			nonEmpty,         // prefix {""}
			empty, esc, term, // "", escaped
			term,       // terminator {""}, within outer
			pNil, term, // nil = outer[1]
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			nonEmpty,    // outer
			empty, term, // {} = outer[0]
			nonEmpty,         // prefix {""}
			empty, esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			nonEmpty,         // outer
			nonEmpty,         // prefix {""}
			empty, esc, term, // "", escaped
			term,        // terminator {""}, within outer
			empty, term, // {} = outer[1]
		}},

		// a complex example
		{"[nil, [a, \"\", xyz], nil, [\"\"], []]", [][]string{
			nil,
			{"a", "", "xyz"},
			nil,
			{""},
			{},
		}, []byte{
			nonEmpty,
			// nil
			pNil, term,
			// {"a", "", "xyz"}, escaped
			nonEmpty,
			nonEmpty, 'a', esc, term,
			empty, esc, term,
			nonEmpty, 'x', 'y', 'z', esc, term,
			term,
			// nil
			pNil, term,
			// {""}, escaped
			nonEmpty, empty, esc, term,
			term,
			// {}
			empty, term,
		}},
	})
}

type sInt []int32

func TestSliceUnderlyingType(t *testing.T) {
	codec := internal.SliceCodec[sInt](int32Codec)
	testCodec(t, codec, []testCase[sInt]{
		{"nil", sInt(nil), []byte{pNil}},
		{"empty", sInt([]int32{}), []byte{empty}},
		{"[0]", sInt([]int32{0}), []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", sInt([]int32{-1}), []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", sInt([]int32{0, 1, -1}), []byte{
			nonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, []int32{})
}
