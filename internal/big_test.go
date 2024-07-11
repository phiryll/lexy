package internal_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func encodeSize(size int64) []byte {
	var buf bytes.Buffer
	internal.Int64Codec.Write(&buf, size)
	return buf.Bytes()
}

func newBigInt(s string) *big.Int {
	var value big.Int
	value.SetString(s, 10)
	return &value
}

func encodeBigInt(s string) []byte {
	value := newBigInt(s)
	var b bytes.Buffer
	if err := internal.BigIntCodec.Write(&b, value); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func TestBigInt(t *testing.T) {
	codec := internal.BigIntCodec
	testCodec(t, codec, []testCase[*big.Int]{
		{"-257", big.NewInt(-257), append(encodeSize(-2),
			[]byte{0xFE, 0xFE}...)},
		{"-256", big.NewInt(-256), append(encodeSize(-2),
			[]byte{0xFE, 0xFF}...)},
		{"-255", big.NewInt(-255), append(encodeSize(-1),
			[]byte{0x00}...)},
		{"-254", big.NewInt(-254), append(encodeSize(-1),
			[]byte{0x01}...)},
		{"-2", big.NewInt(-2), append(encodeSize(-1),
			[]byte{0xFD}...)},
		{"-1", big.NewInt(-1), append(encodeSize(-1),
			[]byte{0xFE}...)},
		{"0", big.NewInt(0), append(encodeSize(0),
			[]byte{}...)},
		{"+1", big.NewInt(1), append(encodeSize(1),
			[]byte{0x01}...)},
		{"+2", big.NewInt(2), append(encodeSize(1),
			[]byte{0x02}...)},
		{"254", big.NewInt(254), append(encodeSize(1),
			[]byte{0xFE}...)},
		{"255", big.NewInt(255), append(encodeSize(1),
			[]byte{0xFF}...)},
		{"256", big.NewInt(256), append(encodeSize(2),
			[]byte{0x01, 0x00}...)},
		{"257", big.NewInt(257), append(encodeSize(2),
			[]byte{0x01, 0x01}...)},
	})

	testCodecRoundTrip(t, codec, []testCase[*big.Int]{
		{"big positive", newBigInt("1234567890123456789012345678901234567890"), nil},
		{"big negative", newBigInt("-1234567890123456789012345678901234567890"), nil},
	})
}

func TestBigIntOrdering(t *testing.T) {
	assert.IsIncreasing(t, [][]byte{
		encodeBigInt("-12345"),
		encodeBigInt("-12344"),
		encodeBigInt("-12343"),
		encodeBigInt("-257"),
		encodeBigInt("-256"),
		encodeBigInt("-255"),
		encodeBigInt("-1"),
		encodeBigInt("0"),
		encodeBigInt("1"),
		encodeBigInt("255"),
		encodeBigInt("256"),
		encodeBigInt("257"),
		encodeBigInt("12343"),
		encodeBigInt("12344"),
		encodeBigInt("12345"),
	})
}

func TestBigFloat(t *testing.T) {

}
