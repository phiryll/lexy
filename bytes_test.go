package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	codec := lexy.BytesCodec[[]byte](true)
	testCodec(t, codec, []testCase[[]byte]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []byte{}, []byte{pNonNil}},
		{"[0]", []byte{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, []byte{0})
}

func TestBytesUnderlyingType(t *testing.T) {
	type header []byte
	codec := lexy.BytesCodec[header](true)
	testCodec(t, codec, []testCase[header]{
		{"nil", header(nil), []byte{pNilFirst}},
		{"empty", header{}, []byte{pNonNil}},
		{"[0]", header{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", header{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
	testCodecFail(t, codec, header{0})
}

func TestBytesNilsLast(t *testing.T) {
	encodeFirst := encoderFor(lexy.BytesCodec[[]byte](true))
	encodeLast := encoderFor(lexy.BytesCodec[[]byte](false))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst([]byte{0}),
		encodeFirst([]byte{0, 0, 0}),
		encodeFirst([]byte{0, 1}),
		encodeFirst([]byte{35}),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast([]byte{0}),
		encodeLast([]byte{0, 0, 0}),
		encodeLast([]byte{0, 1}),
		encodeLast([]byte{35}),
		encodeLast(nil),
	})
}
