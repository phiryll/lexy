package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestPointerInt32(t *testing.T) {
	valueCodec := internal.Int32Codec
	codec := internal.NewPointerCodec[int32](valueCodec)
	testCodec[*int32](t, codec, []testCase[*int32]{
		{"nil", nil, []byte(nil)},
		{"*0", ptr(int32(0)), []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail[*int32](t, codec, ptr(int32(0)))
}

func TestPointerString(t *testing.T) {
	stringCodec := internal.StringCodec
	codec := internal.NewPointerCodec[string](stringCodec)
	testCodec[*string](t, codec, []testCase[*string]{
		{"nil", nil, []byte(nil)},
		{"*empty", ptr(""), []byte{nonEmpty, empty}},
		{"*abc", ptr("abc"), []byte{nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail[*string](t, codec, ptr("abc"))
}

func TestPointerPointerString(t *testing.T) {
	stringCodec := internal.StringCodec
	pointerCodec := internal.NewPointerCodec[string](stringCodec)
	codec := internal.NewPointerCodec[*string](pointerCodec)
	testCodec[**string](t, codec, []testCase[**string]{
		{"nil", nil, []byte(nil)},
		{"*nil", ptr((*string)(nil)), []byte{nonEmpty}},
		{"**empty", ptr(ptr("")), []byte{nonEmpty, nonEmpty, empty}},
		{"**abc", ptr(ptr("abc")), []byte{nonEmpty, nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail[**string](t, codec, ptr(ptr("abc")))
}

func TestPointerSliceInt32(t *testing.T) {
	int32Codec := internal.Int32Codec
	sliceCodec := internal.NewSliceCodec[int32](int32Codec)
	codec := internal.NewPointerCodec[[]int32](sliceCodec)
	testCodec[*[]int32](t, codec, []testCase[*[]int32]{
		{"nil", nil, []byte(nil)},
		{"*nil", ptr([]int32(nil)), []byte{nonEmpty}},
		{"*[]", &[]int32{}, []byte{nonEmpty, empty}},
		{"*[0, 1, -1]", &[]int32{0, 1, -1}, []byte{
			nonEmpty,
			nonEmpty,
			0x80, esc, 0x00, esc, 0x00, esc, 0x00, del,
			0x80, esc, 0x00, esc, 0x00, esc, 0x01, del,
			0x7F, 0xFF, 0xFF, 0xFF,
		}},
	})
	testCodecFail[*[]int32](t, codec, &[]int32{})
}
