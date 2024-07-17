package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestSliceInt32(t *testing.T) {
	codec := internal.MakeSliceCodec[[]int32](int32Codec)
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
	codec := internal.MakeSliceCodec[[]string](stringCodec)
	testCodec(t, codec, []testCase[[]string]{
		{"nil", nil, []byte{pNil}},
		{"empty", []string{}, []byte{empty}},
		{"[\"\"]", []string{""}, []byte{nonEmpty, empty, del}},
		{"[a]", []string{"a"}, []byte{nonEmpty, nonEmpty, 'a', del}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			nonEmpty,
			nonEmpty, 'a', del,
			empty, del,
			nonEmpty, 'x', 'y', 'z', del,
		}},
	})
	testCodecFail(t, codec, []string{})
}

func TestSlicePtrString(t *testing.T) {
	pointerCodec := internal.MakePointerCodec[*string](stringCodec)
	codec := internal.MakeSliceCodec[[]*string](pointerCodec)
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{pNil}},
		{"empty", []*string{}, []byte{empty}},
		{"[nil]", []*string{nil}, []byte{nonEmpty, pNil, del}},
		{"[*a]", []*string{ptr("a")}, []byte{nonEmpty, nonEmpty, nonEmpty, 'a', del}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			nonEmpty,
			nonEmpty, nonEmpty, 'a', del,
			pNil, del,
			nonEmpty, empty, del,
			nonEmpty, nonEmpty, 'x', 'y', 'z', del,
		}},
	})
	testCodecFail(t, codec, []*string{})
}

func TestSliceSliceInt32(t *testing.T) {
	sliceCodec := internal.MakeSliceCodec[[]int32](int32Codec)
	codec := internal.MakeSliceCodec[[][]int32](sliceCodec)
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte{pNil}},
		{"[]", [][]int32{}, []byte{empty}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{nonEmpty, pNil, del}},
		{"[[]]", [][]int32{{}}, []byte{nonEmpty, empty, del}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			// unescaped delimiters are the top level are on separate lines for clarity
			nonEmpty,
			// nil
			pNil, del,
			// {0, 1, -1}, escaped
			nonEmpty,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			del,
			// {}
			empty,
			del,
			// {-2, -3}, escaped
			nonEmpty,
			0x7F, 0xFF, 0xFF, 0xFE,
			0x7F, 0xFF, 0xFF, 0xFD,
			del,
			// nil
			pNil, del,
			// nil
			pNil, del,
		}},
	})
	testCodecFail(t, codec, [][]int32{})
}

func TestSliceSliceString(t *testing.T) {
	sliceCodec := internal.MakeSliceCodec[[]string](stringCodec)
	codec := internal.MakeSliceCodec[[][]string](sliceCodec)

	testCodec(t, codec, []testCase[[][]string]{
		// unescaped delimiters are the top level are on separate lines for clarity
		{"nil", nil, []byte{pNil}},
		{"[]", [][]string{}, []byte{empty}},
		{"[nil]", [][]string{[]string(nil)}, []byte{nonEmpty, pNil, del}},
		{"[[]]", [][]string{{}}, []byte{nonEmpty, empty, del}},
		{"[[\"\"]]", [][]string{{""}}, []byte{
			nonEmpty,        // prefix outer
			nonEmpty,        // prefix {""}
			empty, esc, del, // "", escaped
			del, // terminator {""}, within outer
		}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			nonEmpty,  // outer
			pNil, del, // nil = outer[0]
			empty, del, // {} = outer[1]
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			nonEmpty,   // outer
			empty, del, // {} = outer[0]
			pNil, del, // nil = outer[1]
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			nonEmpty,  // outer
			pNil, del, // nil = outer[0]
			nonEmpty,        // prefix {""}
			empty, esc, del, // "", escaped
			del, // terminator {""}, within outer
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			nonEmpty,        // outer
			nonEmpty,        // prefix {""}
			empty, esc, del, // "", escaped
			del,       // terminator {""}, within outer
			pNil, del, // nil = outer[1]
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			nonEmpty,   // outer
			empty, del, // {} = outer[0]
			nonEmpty,        // prefix {""}
			empty, esc, del, // "", escaped
			del, // terminator {""}, within outer
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			nonEmpty,        // outer
			nonEmpty,        // prefix {""}
			empty, esc, del, // "", escaped
			del,        // terminator {""}, within outer
			empty, del, // {} = outer[1]
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
			pNil, del,
			// {"a", "", "xyz"}, escaped
			nonEmpty,
			nonEmpty, 'a', esc, del,
			empty, esc, del,
			nonEmpty, 'x', 'y', 'z', esc, del,
			del,
			// nil
			pNil, del,
			// {""}, escaped
			nonEmpty, empty, esc, del,
			del,
			// {}
			empty, del,
		}},
	})
}

type sInt []int32

func TestSliceUnderlyingType(t *testing.T) {
	codec := internal.MakeSliceCodec[sInt](int32Codec)
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
