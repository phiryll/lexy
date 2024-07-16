package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestPointerInt32(t *testing.T) {
	codec := internal.MakePointerCodec[*int32](int32Codec)
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte(nil)},
		{"*0", ptr(int32(0)), []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, ptr(int32(0)))
}

func TestPointerString(t *testing.T) {
	codec := internal.MakePointerCodec[*string](stringCodec)
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte(nil)},
		{"*empty", ptr(""), []byte{nonEmpty, empty}},
		{"*abc", ptr("abc"), []byte{nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr("abc"))
}

func TestPointerPointerString(t *testing.T) {
	pointerCodec := internal.MakePointerCodec[*string](stringCodec)
	codec := internal.MakePointerCodec[**string](pointerCodec)
	testCodec(t, codec, []testCase[**string]{
		{"nil", nil, []byte(nil)},
		{"*nil", ptr((*string)(nil)), []byte{nonEmpty}},
		{"**empty", ptr(ptr("")), []byte{nonEmpty, nonEmpty, empty}},
		{"**abc", ptr(ptr("abc")), []byte{nonEmpty, nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr(ptr("abc")))
}

func TestPointerSliceInt32(t *testing.T) {
	sliceCodec := internal.MakeSliceCodec[[]int32](int32Codec)
	codec := internal.MakePointerCodec[*[]int32](sliceCodec)
	testCodec(t, codec, []testCase[*[]int32]{
		{"nil", nil, []byte(nil)},
		{"*nil", ptr([]int32(nil)), []byte{nonEmpty}},
		{"*[]", &[]int32{}, []byte{nonEmpty, empty}},
		{"*[0, 1, -1]", &[]int32{0, 1, -1}, []byte{
			nonEmpty,
			nonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, &[]int32{})
}

type pInt *int32

func TestPointerUnderlyingType(t *testing.T) {
	codec := internal.MakePointerCodec[pInt](int32Codec)
	testCodec(t, codec, []testCase[pInt]{
		{"nil", pInt(nil), []byte(nil)},
		{"*0", pInt(ptr(int32(0))), []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", pInt(ptr(int32(-1))), []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, pInt(ptr(int32(0))))
}
