package internal_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// just making map codec declarations terser
var (
	sCodec       = stringCodec
	iCodec       = int32Codec
	sliceCodec   = internal.SliceCodec[[]string](sCodec, true)
	pointerCodec = internal.PointerCodec[*string](sCodec, true)
)

func testBasicMap[M ~map[string]int32](t *testing.T, codec internal.Codec[M]) {
	// at most one key so order does not matter
	testCodec(t, codec, []testCase[M]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", M{}, []byte{pEmpty}},
		{"{a:0}", M{"a": 0}, []byte{
			pNonEmpty,
			pNonEmpty, 'a', term,
			0x80, 0x00, 0x00, 0x00,
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

func TestMapInt(t *testing.T) {
	testBasicMap(t, internal.MapCodec[map[string]int32](sCodec, iCodec, true))
}

type mStringInt map[string]int32

func TestMapUnderlyingType(t *testing.T) {
	testBasicMap(t, internal.MapCodec[mStringInt](sCodec, iCodec, true))
}

func TestMapSlice(t *testing.T) {
	codec := internal.MapCodec[map[string][]string](sCodec, sliceCodec, true)
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

func TestMapPointerPointer(t *testing.T) {
	// Unfortunately, comparing pointers does not compare what they're pointing to.
	// Instead, we'll dump the pointees into a new map and compare that.
	codec := internal.MapCodec[map[*string]*string](pointerCodec, pointerCodec, true)
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

func TestMapNilsLast(t *testing.T) {
	// Maps are randomly ordered, so we can only test nil/empty/non-empty.
	encodeFirst := encoderFor(internal.MapCodec[map[string]int32](sCodec, iCodec, true))
	encodeLast := encoderFor(internal.MapCodec[map[string]int32](sCodec, iCodec, false))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst(map[string]int32{}),
		encodeFirst(map[string]int32{"a": 0}),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast(map[string]int32{}),
		encodeLast(map[string]int32{"a": 0}),
		encodeLast(nil),
	})
}
