package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestPointerInt32(t *testing.T) {
	codec := toCodec(lexy.PointerTo(lexy.Int32()))
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*0", ptr(int32(0)), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, ptr(int32(0)))
}

func TestPointerString(t *testing.T) {
	codec := toCodec(lexy.PointerTo(lexy.String()))
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(""), []byte{pNonNil}},
		{"*abc", ptr("abc"), []byte{pNonNil, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr("abc"))
}

func TestPointerPointerString(t *testing.T) {
	codec := toCodec(lexy.PointerTo(toCodec(lexy.PointerTo(lexy.String()))))
	testCodec(t, codec, []testCase[**string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*nil", ptr((*string)(nil)), []byte{pNonNil, pNilFirst}},
		{"**empty", ptr(ptr("")), []byte{pNonNil, pNonNil}},
		{"**abc", ptr(ptr("abc")), []byte{pNonNil, pNonNil, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr(ptr("abc")))
}

func TestPointerSliceInt32(t *testing.T) {
	codec := toCodec(lexy.PointerTo(toCodec(lexy.SliceOf(lexy.Int32()))))
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
	testCodecFail(t, codec, &[]int32{})
}

func TestPointerNilsLast(t *testing.T) {
	encodeFirst := encoderFor(toCodec(lexy.PointerTo(lexy.String())))
	encodeLast := encoderFor(toCodec(lexy.PointerTo(lexy.String()).NilsLast()))
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
	codec := toCodec(lexy.MakePointerTo[pInt](lexy.Int32()))
	testCodec(t, codec, []testCase[pInt]{
		{"nil", pInt(nil), []byte{pNilFirst}},
		{"*0", pInt(ptr(int32(0))), []byte{pNonNil, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", pInt(ptr(int32(-1))), []byte{pNonNil, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, pInt(ptr(int32(0))))
}
