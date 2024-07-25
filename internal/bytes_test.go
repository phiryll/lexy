package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestBytes(t *testing.T) {
	codec := internal.BytesCodec[[]byte]()
	testCodec(t, codec, []testCase[[]byte]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []byte{}, []byte{pEmpty}},
		{"[0]", []byte{0}, []byte{pNonEmpty, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{pNonEmpty, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, []byte{0})
}

func TestBytesUnderlyingType(t *testing.T) {
	type header []byte
	codec := internal.BytesCodec[header]()
	testCodec(t, codec, []testCase[header]{
		{"nil", header(nil), []byte{pNilFirst}},
		{"empty", header{}, []byte{pEmpty}},
		{"[0]", header{0}, []byte{pNonEmpty, 0x00}},
		{"[1, 2, 3]", header{1, 2, 3}, []byte{pNonEmpty, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, header{0})
}
