package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestBytes(t *testing.T) {
	codec := internal.MakeBytesCodec[[]byte]()
	testCodec(t, codec, []testCase[[]byte]{
		{"nil", nil, []byte{pNil}},
		{"empty", []byte{}, []byte{empty}},
		{"[0]", []byte{0}, []byte{nonEmpty, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{nonEmpty, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, []byte{0})
}

func TestBytesUnderlyingType(t *testing.T) {
	type header []byte
	codec := internal.MakeBytesCodec[header]()
	testCodec(t, codec, []testCase[header]{
		{"nil", header(nil), []byte{pNil}},
		{"empty", header{}, []byte{empty}},
		{"[0]", header{0}, []byte{nonEmpty, 0x00}},
		{"[1, 2, 3]", header{1, 2, 3}, []byte{nonEmpty, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, header{0})
}
