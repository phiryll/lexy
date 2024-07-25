package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestPointerInt32(t *testing.T) {
	codec := internal.PointerCodec[*int32](int32Codec)
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*0", ptr(int32(0)), []byte{pNonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{pNonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, ptr(int32(0)))
}

func TestPointerString(t *testing.T) {
	codec := internal.PointerCodec[*string](stringCodec)
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(""), []byte{pNonEmpty, pEmpty}},
		{"*abc", ptr("abc"), []byte{pNonEmpty, pNonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr("abc"))
}

func TestPointerPointerString(t *testing.T) {
	pointerCodec := internal.PointerCodec[*string](stringCodec)
	codec := internal.PointerCodec[**string](pointerCodec)
	testCodec(t, codec, []testCase[**string]{
		{"nil", nil, []byte{pNilFirst}},
		{"*nil", ptr((*string)(nil)), []byte{pNonEmpty, pNilFirst}},
		{"**empty", ptr(ptr("")), []byte{pNonEmpty, pNonEmpty, pEmpty}},
		{"**abc", ptr(ptr("abc")), []byte{pNonEmpty, pNonEmpty, pNonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr(ptr("abc")))
}

func TestPointerSliceInt32(t *testing.T) {
	sliceCodec := internal.SliceCodec[[]int32](int32Codec)
	codec := internal.PointerCodec[*[]int32](sliceCodec)
	testCodec(t, codec, []testCase[*[]int32]{
		{"nil", nil, []byte{pNilFirst}},
		{"*nil", ptr([]int32(nil)), []byte{pNonEmpty, pNilFirst}},
		{"*[]", &[]int32{}, []byte{pNonEmpty, pEmpty}},
		{"*[0, 1, -1]", &[]int32{0, 1, -1}, []byte{
			pNonEmpty,
			pNonEmpty,
			0x80, 0x00, 0x00, 0x00,
			0x80, 0x00, 0x00, 0x01,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail(t, codec, &[]int32{})
}

type pInt *int32

func TestPointerUnderlyingType(t *testing.T) {
	codec := internal.PointerCodec[pInt](int32Codec)
	testCodec(t, codec, []testCase[pInt]{
		{"nil", pInt(nil), []byte{pNilFirst}},
		{"*0", pInt(ptr(int32(0))), []byte{pNonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", pInt(ptr(int32(-1))), []byte{pNonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, pInt(ptr(int32(0))))
}
