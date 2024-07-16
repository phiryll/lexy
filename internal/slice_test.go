package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestSliceInt32(t *testing.T) {
	codec := internal.MakeSliceCodec[[]int32](int32Codec)
	testCodec(t, codec, []testCase[[]int32]{
		{"nil", nil, []byte(nil)},
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
		{"nil", nil, []byte(nil)},
		{"empty", []string{}, []byte{empty}},
		{"[\"\"]", []string{""}, []byte{nonEmpty, empty}},
		{"[a]", []string{"a"}, []byte{nonEmpty, nonEmpty, 'a'}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			nonEmpty,
			nonEmpty, 'a', del,
			empty, del,
			nonEmpty, 'x', 'y', 'z',
		}},
	})
	testCodecFail(t, codec, []string{})
}

func TestSlicePtrString(t *testing.T) {
	pointerCodec := internal.MakePointerCodec[*string](stringCodec)
	codec := internal.MakeSliceCodec[[]*string](pointerCodec)
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte(nil)},
		{"empty", []*string{}, []byte{empty}},
		{"[nil]", []*string{nil}, []byte{nonEmpty}},
		{"[*a]", []*string{ptr("a")}, []byte{nonEmpty, nonEmpty, nonEmpty, 'a'}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			nonEmpty,
			nonEmpty, nonEmpty, 'a', del,
			del,
			nonEmpty, empty, del,
			nonEmpty, nonEmpty, 'x', 'y', 'z',
		}},
	})
	testCodecFail(t, codec, []*string{})
}

func TestSliceSliceInt32(t *testing.T) {
	sliceCodec := internal.MakeSliceCodec[[]int32](int32Codec)
	codec := internal.MakeSliceCodec[[][]int32](sliceCodec)
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte(nil)},
		{"[]", [][]int32{}, []byte{empty}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{nonEmpty}},
		{"[[]]", [][]int32{{}}, []byte{nonEmpty, empty}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			// unescaped delimiters are the top level are on separate lines for clarity
			nonEmpty,
			// nil
			del,
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
			del,
			// nil
		}},
	})
	testCodecFail(t, codec, [][]int32{})
}

func TestSliceSliceString(t *testing.T) {
	sliceCodec := internal.MakeSliceCodec[[]string](stringCodec)
	codec := internal.MakeSliceCodec[[][]string](sliceCodec)

	testCodec(t, codec, []testCase[[][]string]{
		// unescaped delimiters are the top level are on separate lines for clarity
		{"nil", nil, []byte(nil)},
		{"[]", [][]string{}, []byte{empty}},
		{"[nil]", [][]string{[]string(nil)}, []byte{nonEmpty}},
		{"[[]]", [][]string{{}}, []byte{nonEmpty, empty}},
		{"[[\"\"]]", [][]string{{""}}, []byte{nonEmpty, nonEmpty, empty}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			nonEmpty,
			del,
			empty,
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			nonEmpty,
			empty,
			del,
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			nonEmpty,
			del,
			nonEmpty, empty,
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			nonEmpty,
			nonEmpty, empty,
			del,
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			nonEmpty,
			empty,
			del,
			nonEmpty, empty,
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			nonEmpty,
			nonEmpty, empty,
			del,
			empty,
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
			del,
			// {"a", "", "xyz"}, escaped
			nonEmpty,
			nonEmpty, 'a', esc, del,
			empty, esc, del,
			nonEmpty, 'x', 'y', 'z',
			del,
			// nil
			del,
			// {""}, escaped
			nonEmpty, empty,
			del,
			// {}
			empty,
		}},
	})
}

type sInt []int32

func TestSliceUnderlyingType(t *testing.T) {
	codec := internal.MakeSliceCodec[sInt](int32Codec)
	testCodec(t, codec, []testCase[sInt]{
		{"nil", sInt(nil), []byte(nil)},
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
