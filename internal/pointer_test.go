package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestPointerInt32(t *testing.T) {
	valueCodec := internal.IntCodec[int32]()
	codec := internal.MakePointerCodec(valueCodec)
	testCodec(t, codec, []testCase[*int32]{
		{"nil", nil, []byte(nil)},
		{"*0", ptr(int32(0)), []byte{nonEmpty, 0x80, 0x00, 0x00, 0x00}},
		{"*-1", ptr(int32(-1)), []byte{nonEmpty, 0x7F, 0xFF, 0xFF, 0xFF}},
	})
	testCodecFail(t, codec, ptr(int32(0)))
}

func TestPointerString(t *testing.T) {
	stringCodec := internal.StringCodec
	codec := internal.MakePointerCodec(stringCodec)
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte(nil)},
		{"*empty", ptr(""), []byte{nonEmpty, empty}},
		{"*abc", ptr("abc"), []byte{nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr("abc"))
}

func TestPointerPointerString(t *testing.T) {
	stringCodec := internal.StringCodec
	pointerCodec := internal.MakePointerCodec(stringCodec)
	codec := internal.MakePointerCodec(pointerCodec)
	testCodec(t, codec, []testCase[**string]{
		{"nil", nil, []byte(nil)},
		{"*nil", ptr((*string)(nil)), []byte{nonEmpty}},
		{"**empty", ptr(ptr("")), []byte{nonEmpty, nonEmpty, empty}},
		{"**abc", ptr(ptr("abc")), []byte{nonEmpty, nonEmpty, nonEmpty, 'a', 'b', 'c'}},
	})
	testCodecFail(t, codec, ptr(ptr("abc")))
}

func TestPointerSliceInt32(t *testing.T) {
	int32Codec := internal.IntCodec[int32]()
	sliceCodec := internal.MakeSliceCodec(int32Codec)
	codec := internal.MakePointerCodec(sliceCodec)
	testCodec(t, codec, []testCase[*[]int32]{
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
	testCodecFail(t, codec, &[]int32{})
}
