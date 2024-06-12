package internal

import (
	"bytes"
	"math"
	"testing"
)

func TestBoolCodec_Read(t *testing.T) {
	testRead[bool](t, UintCodec[bool]{}, []readTestCase[bool]{
		{"false", []byte{0}, false, false},
		{"true", []byte{1}, true, false},
		{"fail", []byte{}, false, true},
	})
}

func TestBoolCodec_Write(t *testing.T) {
	testWrite[bool](t, UintCodec[bool]{}, []writeTestCase[bool]{
		{"false", &bytes.Buffer{}, false, []byte{0}, false},
		{"true", &bytes.Buffer{}, true, []byte{1}, false},
		{"fail", failWriter{}, true, nil, true},
	})
}

func TestUint8Codec_Read(t *testing.T) {
	testRead[uint8](t, UintCodec[uint8]{}, []readTestCase[uint8]{
		{"0x00", []byte{0x00}, 0x00, false},
		{"0x01", []byte{0x01}, 0x01, false},
		{"0x7F", []byte{0x7F}, 0x7F, false},
		{"0x80", []byte{0x80}, 0x80, false},
		{"0xFF", []byte{0xFF}, 0xFF, false},
		{"fail", []byte{}, 0, true},
	})
}

func TestUint8Codec_Write(t *testing.T) {
	testWrite[uint8](t, UintCodec[uint8]{}, []writeTestCase[uint8]{
		{"0x00", &bytes.Buffer{}, 0x00, []byte{0x00}, false},
		{"0x01", &bytes.Buffer{}, 0x01, []byte{0x01}, false},
		{"0x7F", &bytes.Buffer{}, 0x7F, []byte{0x7F}, false},
		{"0x80", &bytes.Buffer{}, 0x80, []byte{0x80}, false},
		{"0xFF", &bytes.Buffer{}, 0xFF, []byte{0xFF}, false},
		{"fail", failWriter{}, 0, nil, true},
	})
}

func TestInt8Codec_Read(t *testing.T) {
	testRead[int8](t, IntCodec[int8]{Mask: math.MinInt8}, []readTestCase[int8]{
		{"min", []byte{0x00}, math.MinInt8, false},
		{"-1", []byte{0x7F}, -1, false},
		{"0", []byte{0x80}, 0, false},
		{"+1", []byte{0x81}, 1, false},
		{"max", []byte{0xFF}, math.MaxInt8, false},
		{"fail", []byte{}, 0, true},
	})
}

func TestInt8Codec_Write(t *testing.T) {
	testWrite[int8](t, IntCodec[int8]{Mask: math.MinInt8}, []writeTestCase[int8]{
		{"min", &bytes.Buffer{}, math.MinInt8, []byte{0x00}, false},
		{"-1", &bytes.Buffer{}, -1, []byte{0x7F}, false},
		{"0", &bytes.Buffer{}, 0, []byte{0x80}, false},
		{"+1", &bytes.Buffer{}, 1, []byte{0x81}, false},
		{"max", &bytes.Buffer{}, math.MaxInt8, []byte{0xFF}, false},
		{"fail", failWriter{}, 0, nil, true},
	})
}
