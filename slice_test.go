package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestSliceInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.Int32())
	testCodec(t, codec, []testCase[[]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []int32{}, []byte{pNonNil}},
		{"[0]", []int32{0}, []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", []int32{-1}, []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", []int32{0, 1, -1}, []byte{
			pNonNil,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
}

func TestCastSliceInt32(t *testing.T) {
	t.Parallel()
	type mySlice []int32
	codec := lexy.CastSliceOf[mySlice](lexy.Int32())
	testCodec(t, codec, []testCase[mySlice]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []int32{}, []byte{pNonNil}},
		{"[0]", []int32{0}, []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", []int32{-1}, []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", []int32{0, 1, -1}, []byte{
			pNonNil,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
}

func TestSliceString(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.String())
	testCodec(t, codec, []testCase[[]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []string{}, []byte{pNonNil}},
		{"[\"\"]", []string{""}, []byte{pNonNil, term}},
		{"[a]", []string{"a"}, []byte{pNonNil, 'a', term}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			pNonNil,
			'a', term,
			term,
			'x', 'y', 'z', term,
		}},
	})
}

func TestCastSliceString(t *testing.T) {
	t.Parallel()
	type mySlice []string
	codec := lexy.CastSliceOf[mySlice](lexy.String())
	testCodec(t, codec, []testCase[mySlice]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []string{}, []byte{pNonNil}},
		{"[\"\"]", []string{""}, []byte{pNonNil, term}},
		{"[a]", []string{"a"}, []byte{pNonNil, 'a', term}},
		{"[a, \"\", xyz]", []string{"a", "", "xyz"}, []byte{
			pNonNil,
			'a', term,
			term,
			'x', 'y', 'z', term,
		}},
	})
}

// Unlike []*string, *uint8 does not require a terminator.
// This tests for the special case of a slice of non-terminated, prefixed elements.
func TestSlicePtrUint8(t *testing.T) {
	t.Parallel()
	p := ptr[uint8]
	codec := lexy.SliceOf(lexy.PointerTo(lexy.Uint8()))
	testCodec(t, codec, []testCase[[]*uint8]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []*uint8{}, []byte{pNonNil}},
		{"[nil]", []*uint8{nil}, []byte{pNonNil, pNilFirst}},
		{"[*12]", []*uint8{p(12)}, []byte{pNonNil, pNonNil, 12}},
		{"[nil, *7, *3, nil, nil, *9, nil]", []*uint8{nil, p(7), p(3), nil, nil, p(9), nil}, []byte{
			pNonNil,
			pNilFirst,
			pNonNil, 7,
			pNonNil, 3,
			pNilFirst,
			pNilFirst,
			pNonNil, 9,
			pNilFirst,
		}},
	})
}

func TestSlicePtrString(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.PointerTo(lexy.String()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []*string{}, []byte{pNonNil}},
		{"[nil]", []*string{nil}, []byte{pNonNil, pNilFirst, term}},
		{"[*a]", []*string{ptr("a")}, []byte{pNonNil, pNonNil, 'a', term}},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, []byte{
			pNonNil,
			pNonNil, 'a', term,
			pNilFirst, term,
			pNonNil, term,
			pNonNil, 'x', 'y', 'z', term,
		}},
	})
}

func TestSliceSliceInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.SliceOf(lexy.Int32()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[[][]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]int32{}, []byte{pNonNil}},
		{"[nil]", [][]int32{[]int32(nil)}, []byte{pNonNil, pNilFirst, term}},
		{"[[]]", [][]int32{{}}, []byte{pNonNil, pNonNil, term}},
		{"[nil, {0, 1, -1}, {}, {-2, -3}, nil, nil]", [][]int32{nil, {0, 1, -1}, {}, {-2, -3}, nil, nil}, []byte{
			pNonNil,
			// nil
			pNilFirst, term,
			// {0, 1, -1}, escaped
			pNonNil,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
			term,
			// {}
			pNonNil,
			term,
			// {-2, -3}, escaped
			pNonNil,
			0x7F, 0xFF, 0xFF, 0xFE,
			0x7F, 0xFF, 0xFF, 0xFD,
			term,
			// nil
			pNilFirst, term,
			// nil
			pNilFirst, term,
		}},
	})
}

func TestSliceSliceString(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.SliceOf(lexy.String()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[[][]string]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", [][]string{}, []byte{pNonNil}},
		{"[nil]", [][]string{[]string(nil)}, []byte{pNonNil, pNilFirst, term}},
		{"[[]]", [][]string{{}}, []byte{pNonNil, pNonNil, term}},
		{"[[\"\"]]", [][]string{{""}}, []byte{
			pNonNil,   // prefix outer
			pNonNil,   // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},

		// pairwise permutations of nil, {}, and {""}
		{"[nil, []]", [][]string{nil, {}}, []byte{
			pNonNil,         // outer
			pNilFirst, term, // nil = outer[0]
			pNonNil, term, // {} = outer[1]
		}},
		{"[[], nil]", [][]string{{}, nil}, []byte{
			pNonNil,       // outer
			pNonNil, term, // {} = outer[0]
			pNilFirst, term, // nil = outer[1]
		}},
		{"[nil, [\"\"]]", [][]string{nil, {""}}, []byte{
			pNonNil,         // outer
			pNilFirst, term, // nil = outer[0]
			pNonNil,   // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], nil]", [][]string{{""}, nil}, []byte{
			pNonNil,   // outer
			pNonNil,   // prefix {""}
			esc, term, // "", escaped
			term,            // terminator {""}, within outer
			pNilFirst, term, // nil = outer[1]
		}},
		{"[[], [\"\"]]", [][]string{{}, {""}}, []byte{
			pNonNil,       // outer
			pNonNil, term, // {} = outer[0]
			pNonNil,   // prefix {""}
			esc, term, // "", escaped
			term, // terminator {""}, within outer
		}},
		{"[[\"\"], []]", [][]string{{""}, {}}, []byte{
			pNonNil,   // outer
			pNonNil,   // prefix {""}
			esc, term, // "", escaped
			term,          // terminator {""}, within outer
			pNonNil, term, // {} = outer[1]
		}},

		// a complex example
		{"[nil, [a, \"\", xyz], nil, [\"\"], []]", [][]string{
			nil,
			{"a", "", "xyz"},
			nil,
			{""},
			{},
		}, []byte{
			pNonNil,
			// nil
			pNilFirst, term,
			// {"a", "", "xyz"}, escaped
			pNonNil,
			'a', esc, term,
			esc, term,
			'x', 'y', 'z', esc, term,
			term,
			// nil
			pNilFirst, term,
			// {""}, escaped
			pNonNil, esc, term,
			term,
			// {}
			pNonNil, term,
		}},
	})
}

func TestSliceNilsLast(t *testing.T) {
	t.Parallel()
	codec := lexy.SliceOf(lexy.Int32())
	assert.True(t, codec.RequiresTerminator())
	testOrdering(t, lexy.NilsLast(codec), []testCase[[]int32]{
		{"[-100, 5]", []int32{-100, 5}, nil},
		{"[0]", []int32{0}, nil},
		{"[0, 0, 0]", []int32{0, 0, 0}, nil},
		{"[0, 1]", []int32{0, 1}, nil},
		{"[35]", []int32{35}, nil},
		{"nil", nil, nil},
	})
}

func TestCastSliceNilsLast(t *testing.T) {
	t.Parallel()
	type mySlice []int32
	codec := lexy.CastSliceOf[mySlice](lexy.Int32())
	assert.True(t, codec.RequiresTerminator())
	testOrdering(t, lexy.NilsLast(codec), []testCase[mySlice]{
		{"[-100, 5]", []int32{-100, 5}, nil},
		{"[0]", []int32{0}, nil},
		{"[0, 0, 0]", []int32{0, 0, 0}, nil},
		{"[0, 1]", []int32{0, 1}, nil},
		{"[35]", []int32{35}, nil},
		{"nil", nil, nil},
	})
}
