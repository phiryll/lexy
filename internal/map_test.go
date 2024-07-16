package internal_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBasicMap(t *testing.T, codec internal.Codec[map[string]int32]) {
	// at most one key so order does not matter
	testCodec(t, codec, []testCase[map[string]int32]{
		{"nil", nil, []byte(nil)},
		{"empty", map[string]int32{}, []byte{empty}},
		{"{a:0}", map[string]int32{"a": 0}, []byte{
			nonEmpty,
			nonEmpty, 'a', del,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00,
		}},
	})
	testCodecFail(t, codec, map[string]int32{})

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
		// only last if testing an ordered map codec
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

// nil => "nil"
func dePointerMap(m map[*string]*string) map[string]string {
	deref := func(p *string) string {
		if p == nil {
			return "nil"
		}
		return *p
	}
	result := make(map[string]string)
	for k, v := range m {
		result[deref(k)] = deref(v)
	}
	return result
}

// So we can test pointer keys.
func testMapPointerPointer(t *testing.T, codec internal.Codec[map[*string]*string]) {
	// Unfortunately, comparing pointers does not compare what they're pointing to.
	// Instead, we'll dump the pointees into a new map and compare that.
	tests := []testCase[map[*string]*string]{
		{"nil map", map[*string]*string(nil), nil},
		{"empty map", map[*string]*string{}, nil},
		{"non-trivial", map[*string]*string{
			ptr("a"): ptr("1"),
			ptr("c"): nil,
			ptr("b"): ptr(""),
			nil:      ptr("2"),
			ptr(""):  ptr("3"),
		}, nil},
		{"nil-nil", map[*string]*string{
			ptr("a"): ptr("1"),
			ptr("c"): nil,
			nil:      nil,
			ptr("b"): ptr(""),
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			err := codec.Write(&b, tt.value)
			require.NoError(t, err)
			got, err := codec.Read(bytes.NewReader(b.Bytes()))
			require.NoError(t, err)
			assert.Equal(t, dePointerMap(tt.value), dePointerMap(got))
		})
	}
}

// just making map codec declarations terser
var (
	sCodec       = internal.StringCodec
	iCodec       = int32Codec
	sliceCodec   = internal.MakeSliceCodec(sCodec)
	pointerCodec = internal.MakePointerCodec(sCodec)
)

func TestMapInt(t *testing.T) {
	testBasicMap(t, internal.MakeMapCodec(sCodec, iCodec))
}

func TestMapSlice(t *testing.T) {
	testMapSliceValue(t, internal.MakeMapCodec(sCodec, sliceCodec))
}

func TestMapPointerPointer(t *testing.T) {
	testMapPointerPointer(t, internal.MakeMapCodec(pointerCodec, pointerCodec))
}

func TestOrderedMapInt(t *testing.T) {
	testBasicMap(t, internal.MakeOrderedMapCodec(sCodec, iCodec))
}

func TestOrderedMapSlice(t *testing.T) {
	testMapSliceValue(t, internal.MakeOrderedMapCodec(sCodec, sliceCodec))
}

func TestOrderedMapPointerPointer(t *testing.T) {
	testMapPointerPointer(t, internal.MakeOrderedMapCodec(pointerCodec, pointerCodec))
}

func TestOrderedMapOrdering(t *testing.T) {
	// There's no way to force a map to iterate in a particular order, so this test might be accidentally working.
	// We can't even test a map that we've found does not iterate in key order,
	// because go randomizes the initial start for map iteration to prevent depending on iteration order.
	// Because of this, this test might not fail on any particular run,
	// but it should absolutely fail on enough repeated runs if the codec isn't working.
	codec := internal.MakeOrderedMapCodec(sCodec, sCodec)
	testCodec(t, codec, []testCase[map[string]string]{
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
