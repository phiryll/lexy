package internal_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBasicMap[M ~map[string]int32](t *testing.T, codec internal.Codec[M]) {
	// at most one key so order does not matter
	testCodec(t, codec, []testCase[M]{
		{"nil", nil, []byte(nil)},
		{"empty", M{}, []byte{empty}},
		{"{a:0}", M{"a": 0}, []byte{
			nonEmpty,
			nonEmpty, 'a', del,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00, del,
		}},
	})
	testCodecFail(t, codec, M{})

	testCodecRoundTrip(t, codec, []testCase[M]{
		{"non-trivial", M{
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
	sCodec       = stringCodec
	iCodec       = int32Codec
	sliceCodec   = internal.MakeSliceCodec[[]string](sCodec)
	pointerCodec = internal.MakePointerCodec[*string](sCodec)
)

func TestMapInt(t *testing.T) {
	testBasicMap(t, internal.MakeMapCodec[map[string]int32](sCodec, iCodec))
}

func TestMapSlice(t *testing.T) {
	testMapSliceValue(t, internal.MakeMapCodec[map[string][]string](sCodec, sliceCodec))
}

func TestMapPointerPointer(t *testing.T) {
	testMapPointerPointer(t, internal.MakeMapCodec[map[*string]*string](pointerCodec, pointerCodec))
}

type mStringInt map[string]int32

func TestMapUnderlyingType(t *testing.T) {
	testBasicMap(t, internal.MakeMapCodec[mStringInt](sCodec, iCodec))
}
