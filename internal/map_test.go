package internal_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
)

func testBasicMap(t *testing.T, codec internal.Codec[map[string]int32]) {
	// at most one key so order does not matter
	testCodec[map[string]int32](t, codec, []testCase[map[string]int32]{
		{"nil", nil, []byte(nil)},
		{"empty", map[string]int32{}, []byte{empty}},
		{"{a:0}", map[string]int32{"a": 0}, []byte{
			nonEmpty,
			nonEmpty, 'a', del,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
		}},
	})
	testCodecFail[map[string]int32](t, codec, map[string]int32{})

	testCodecRoundTrip(t, codec, []testCase[map[string]int32]{
		{"non-trivial", map[string]int32{
			"a": 0,
			"b": -1,
			"":  1000,
			"c": math.MaxInt32,
			"d": math.MinInt32,
		}, nil},
	})
}

// So we can test nil and empty values.
func testMapSliceValue(t *testing.T, codec internal.Codec[map[string][]string]) {
	testCodecRoundTrip(t, codec, []testCase[map[string][]string]{
		{"nil map", map[string][]string(nil), nil},
		{"empty map", map[string][]string{}, nil},
		{"nil last", map[string][]string{
			"a": {"x", "y", "zq"},
			"b": {},
			"":  {"p", "q"},
			"c": nil,
		}, nil},
		{"empty last", map[string][]string{
			"a": {"x", "y", "zq"},
			"b": nil,
			"":  {"p", "q"},
			"c": {},
		}, nil},
	})
}

// just making map codec declarations terser
var (
	sCodec     = internal.StringCodec
	iCodec     = internal.Int32Codec
	sliceCodec = internal.NewSliceCodec(sCodec)
)

func TestMapInt(t *testing.T) {
	testBasicMap(t, internal.NewMapCodec(sCodec, iCodec))
}

func TestMapSlice(t *testing.T) {
	testMapSliceValue(t, internal.NewMapCodec(sCodec, sliceCodec))
}

func TestOrderedMapInt(t *testing.T) {
	testBasicMap(t, internal.NewOrderedMapCodec(sCodec, iCodec))
}

func TestOrderedMapSlice(t *testing.T) {
	testMapSliceValue(t, internal.NewOrderedMapCodec(sCodec, sliceCodec))
}

func TestOrderedMapOrdering(t *testing.T) {
	// There's no way to force a map to iterate in a particular order, so this test might be accidentally working.
	// We can't even test a map that we've found does not iterate in key order,
	// because go randomizes the initial start for map iteration to prevent depending on iteration order.
	// Because of this, this test might not fail on any particular run,
	// but it should absolutely fail on enough repeated runs if the codec isn't working.
	codec := internal.NewOrderedMapCodec(sCodec, sCodec)
	testCodec[map[string]string](t, codec, []testCase[map[string]string]{
		{"nil", nil, []byte(nil)},
		{"empty", map[string]string{}, []byte{empty}},
		{"non-empty", map[string]string{
			"b": "3",
			"":  "1",
			"c": "4",
			"a": "2",
		}, []byte{
			nonEmpty,
			empty, del, nonEmpty, '1', del,
			nonEmpty, 'a', del, nonEmpty, '2', del,
			nonEmpty, 'b', del, nonEmpty, '3', del,
			nonEmpty, 'c', del, nonEmpty, '4',
		}},
	})
}
