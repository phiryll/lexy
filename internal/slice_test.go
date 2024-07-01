package internal_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestSliceInt32(t *testing.T) {
	elementCodec := internal.IntCodec[int32]{Mask: math.MinInt32}
	codec := internal.NewSliceCodec[int32](elementCodec)
	testCodec[[]int32](t, codec, []testCase[[]int32]{
		{"nil", nil, []byte(nil)},
		{"empty", []int32{}, []byte{zero}},
		{"[0]", []int32{0}, []byte{nonZero, 0x80, esc, 0x00, esc, 0x00, esc, 0x00}},
		{"[-1]", []int32{-1}, []byte{nonZero, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", []int32{0, 1, -1}, []byte{
			nonZero,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00, del,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01, del,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail[[]int32](t, codec, []int32{})
}

func TestSliceString(t *testing.T) {
	stringCodec := internal.StringCodec{}
	codec := internal.NewSliceCodec[string](stringCodec)
	testCodec[[]string](t, codec, []testCase[[]string]{
		{"nil", nil, []byte(nil)},
		{"empty", []string{}, []byte{zero}},
		{"[\"\"]", []string{""}, []byte{nonZero, zero}},
		{"[a]", []string{"a"}, []byte{nonZero, nonZero, 'a'}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			nonZero,
			nonZero, 'a', del,
			zero, del,
			nonZero, 'x', 'y', 'z',
		}},
	})
	testCodecFail[[]string](t, codec, []string{})
}

func TestSliceSliceInt32(t *testing.T) {
	int32Codec := internal.IntCodec[int32]{Mask: math.MinInt32}
	sliceCodec := internal.NewSliceCodec[int32](int32Codec)
	codec := internal.NewSliceCodec[[]int32](sliceCodec)
	testCodec[[][]int32](t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte(nil)},
		{"[]", [][]int32{}, []byte{zero}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{nonZero}},
		{"[[]]", [][]int32{{}}, []byte{nonZero, zero}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			// unescaped delimiters are the top level are on separate lines for clarity
			nonZero,
			// nil
			del,
			// {0, 1, -1}, escaped
			nonZero,
			0x80, esc, esc, esc, 0x00, esc, esc, esc, 0x00, esc, esc, esc, 0x00, esc, del,
			0x80, esc, esc, esc, 0x00, esc, esc, esc, 0x00, esc, esc, esc, 0x01, esc, del,
			0x7F, 0xFF, 0xFF, 0xFF,
			del,
			// {}
			zero,
			del,
			// {-2, -3}, escaped
			nonZero,
			0x7F, 0xFF, 0xFF, 0xFE, esc, del,
			0x7F, 0xFF, 0xFF, 0xFD,
			del,
			// nil
			del,
			// nil
		}},
	})
	testCodecFail[[][]int32](t, codec, [][]int32{})
}

func TestSliceSliceString(t *testing.T) {
	stringCodec := internal.StringCodec{}
	sliceCodec := internal.NewSliceCodec[string](stringCodec)
	codec := internal.NewSliceCodec[[]string](sliceCodec)

	testCodec[[][]string](t, codec, []testCase[[][]string]{
		// unescaped delimiters are the top level are on separate lines for clarity
		{"nil", nil, []byte(nil)},
		{"[]", [][]string{}, []byte{zero}},
		{"[nil]", [][]string{[]string(nil)}, []byte{nonZero}},
		{"[[]]", [][]string{{}}, []byte{nonZero, zero}},
		{"[[\"\"]]", [][]string{{""}}, []byte{nonZero, nonZero, zero}},

		// pairwise permutations of nil, [], and [""]
		{"[nil, []]", [][]string{nil, {}}, []byte{
			nonZero,
			del,
			zero,
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			nonZero,
			zero,
			del,
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			nonZero,
			del,
			nonZero, zero,
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			nonZero,
			nonZero, zero,
			del,
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			nonZero,
			zero,
			del,
			nonZero, zero,
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			nonZero,
			nonZero, zero,
			del,
			zero,
		}},

		// a complex example
		{"[nil, [a, \"\", xyz], nil, [\"\"], []]", [][]string{
			nil,
			{"a", "", "xyz"},
			nil,
			{""},
			{},
		}, []byte{
			nonZero,
			// nil
			del,
			// {"a", "", "xyz"}, escaped
			nonZero,
			nonZero, 'a', esc, del,
			zero, esc, del,
			nonZero, 'x', 'y', 'z',
			del,
			// nil
			del,
			// {""}, escaped
			nonZero, zero,
			del,
			// {}
			zero,
		}},
	})
}
