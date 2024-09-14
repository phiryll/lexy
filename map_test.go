package lexy_test

import (
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func testBasicMap[M ~map[string]int32](t *testing.T, codec lexy.Codec[M]) {
	testBasicMapWithPrefix(t, pNilFirst, codec)
}

func testBasicMapWithPrefix[M ~map[string]int32](t *testing.T, nilPrefix byte, codec lexy.Codec[M]) {
	assert.True(t, codec.RequiresTerminator())

	// at most one key so order does not matter
	testCodec(t, codec, []testCase[M]{
		{"nil", nil, []byte{nilPrefix}},
		{"empty", M{}, []byte{pNonNil}},
		{"{a:0}", M{"a": 0}, []byte{
			pNonNil,
			'a', term,
			0x80, 0x00, 0x00, 0x00,
		}},
	})

	testVaryingCodec(t, codec, []testCase[M]{
		{"non-trivial", M{
			"a": 0,
			"b": -1,
			"":  1000,
			"c": math.MaxInt32,
			"d": math.MinInt32,
		}, nil},
	})
}

// Derefs all the pointers, with nil => "nil".
func dePointerMap(m map[*string]*string) map[string]string {
	deref := func(p *string) string {
		if p == nil {
			return "nil"
		}
		return *p
	}
	result := map[string]string{}
	for k, v := range m {
		result[deref(k)] = deref(v)
	}
	return result
}

func TestMapInt(t *testing.T) {
	t.Parallel()
	testBasicMap(t, lexy.MapOf(lexy.String(), lexy.Int32()))
}

func TestCastMapInt(t *testing.T) {
	t.Parallel()
	type myMap map[string]int32
	testBasicMap(t, lexy.CastMapOf[myMap](lexy.String(), lexy.Int32()))
}

func TestMapSlice(t *testing.T) {
	t.Parallel()
	codec := lexy.MapOf(lexy.String(), lexy.SliceOf(lexy.String()))
	assert.True(t, codec.RequiresTerminator())
	testVaryingCodec(t, codec, []testCase[map[string][]string]{
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
	t.Parallel()
	// Unfortunately, comparing pointers does not compare what they're pointing to.
	// Instead, we'll dump the referents into a new map and compare that.
	pointerCodec := lexy.PointerTo(lexy.String())
	codec := lexy.MapOf(pointerCodec, pointerCodec)
	assert.True(t, codec.RequiresTerminator())
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := codec.Append(nil, tt.value)
			got, _ := codec.Get(buf)
			assert.Equal(t, dePointerMap(tt.value), dePointerMap(got))
		})
	}
}

func TestMapNilsLast(t *testing.T) {
	t.Parallel()
	// Maps are randomly ordered, so we can only test nil/non-nil.
	codec := lexy.MapOf(lexy.String(), lexy.Int32())
	testOrdering(t, lexy.NilsLast(codec), []testCase[map[string]int32]{
		{"empty", map[string]int32{}, nil},
		{"non-empty", map[string]int32{"a": 0}, nil},
		{"nil", nil, nil},
	})
}

func TestCastMapNilsLast(t *testing.T) {
	t.Parallel()
	// Maps are randomly ordered, so we can only test nil/non-nil.
	type myMap map[string]int32
	codec := lexy.CastMapOf[myMap](lexy.String(), lexy.Int32())
	testOrdering(t, lexy.NilsLast(codec), []testCase[myMap]{
		{"empty", map[string]int32{}, nil},
		{"non-empty", map[string]int32{"a": 0}, nil},
		{"nil", nil, nil},
	})
}
