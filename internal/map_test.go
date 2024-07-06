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
