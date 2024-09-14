package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestBytesUnderlyingType(t *testing.T) {
	t.Parallel()
	type header []byte
	codec := lexy.NilsLast(lexy.CastBytes[header]())
	testCodec(t, codec, []testCase[header]{
		{"nil", header(nil), []byte{pNilLast}},
		{"empty", header{}, []byte{pNonNil}},
		{"[0]", header{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", header{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
}

func TestMapUnderlyingType(t *testing.T) {
	t.Parallel()
	type mStringInt map[string]int32
	testBasicMapWithPrefix(t, pNilLast, lexy.NilsLast(lexy.CastMapOf[mStringInt](lexy.String(), lexy.Int32())))
}

func TestSliceUnderlyingType(t *testing.T) {
	t.Parallel()
	type sInt []int32
	codec := lexy.NilsLast(lexy.CastSliceOf[sInt](lexy.Int32()))
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[sInt]{
		{"nil", sInt(nil), []byte{pNilLast}},
		{"empty", sInt([]int32{}), []byte{pNonNil}},
		{"[0]", sInt([]int32{0}), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"[-1]", sInt([]int32{-1}), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"[0, 1, -1]", sInt([]int32{0, 1, -1}), []byte{
			pNonNil,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
}
