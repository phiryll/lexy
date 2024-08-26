package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestPointerInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.Int32())
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*0", ptr(int32(0)), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
}

func TestPointerString(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.String())
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(""), []byte{pNonNil}},
		{"*abc", ptr("abc"), []byte{pNonNil, 'a', 'b', 'c'}},
	})
}

func TestPointerPointerString(t *testing.T) {
	t.Parallel()
	codec := lexy.PointerTo(lexy.PointerTo(lexy.String()))
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
	encodeFirst := encoderFor(lexy.PointerTo(lexy.String()))
	encodeLast := encoderFor(lexy.NilsLast(lexy.PointerTo(lexy.String())))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst(ptr("")),
		encodeFirst(ptr("abc")),
		encodeFirst(ptr("xyz")),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast(ptr("")),
		encodeLast(ptr("abc")),
		encodeLast(ptr("xyz")),
		encodeLast(nil),
	})
}

type pInt *int32

func TestPointerUnderlyingType(t *testing.T) {
	t.Parallel()
	codec := lexy.NilsLast(lexy.CastPointerTo[pInt](lexy.Int32()))
	testCodec(t, codec, []testCase[pInt]{
		{"nil", pInt(nil), []byte{pNilLast}},
		{"*0", pInt(ptr(int32(0))), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", pInt(ptr(int32(-1))), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
}
