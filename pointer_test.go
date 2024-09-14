package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestPointerInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.Int32())
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*0", ptr(int32(0)), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
}

func TestPointerString(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.String())
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(""), []byte{pNonNil}},
		{"*abc", ptr("abc"), []byte{pNonNil, 'a', 'b', 'c'}},
	})
}

func TestPointerPointerString(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.PointerTo(lexy.String()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[**string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*nil", ptr((*string)(nil)), []byte{pNonNil, pNilFirst}},
		{"**empty", ptr(ptr("")), []byte{pNonNil, pNonNil}},
		{"**abc", ptr(ptr("abc")), []byte{pNonNil, pNonNil, 'a', 'b', 'c'}},
	})
}

func TestPointerSliceInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.SliceOf(lexy.Int32()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[*[]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*nil", ptr([]int32(nil)), []byte{pNonNil, pNilFirst}},
		{"*[]", &[]int32{}, []byte{pNonNil, pNonNil}},
		{"*[0, 1, -1]", &[]int32{0, 1, -1}, []byte{
			pNonNil,
			pNonNil,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
}

func TestPointerNilsLast(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.String())
	assert.True(t, codec.RequiresTerminator())
	testOrdering(t, lexy.NilsLast(codec), []testCase[*string]{
		{"*empty", ptr(""), nil},
		{"*abc", ptr("abc"), nil},
		{"*xyz", ptr("xyz"), nil},
		{"nil", nil, nil},
	})
}
