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
	internal.IntCodec[int64]().Write(&buf, size)
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

func newBigFloat(f float64, shift int, prec uint) *big.Float {
	value := big.NewFloat(f)
	value.SetPrec(prec)
	value.SetMantExp(value, shift)
	return value
}

func encodeBigFloat(value *big.Float) []byte {
	var b bytes.Buffer
	if err := internal.BigFloatCodec.Write(&b, value); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func TestBigFloat(t *testing.T) {
	var negInf, posInf, negZero, posZero big.Float
	negInf.SetInf(true)
	posInf.SetInf(false)
	negZero.Neg(&negZero)
	assert.True(t, negZero.Signbit())
	assert.False(t, posZero.Signbit())
	assert.Equal(t, 0, negZero.Cmp(&posZero))
	assert.NotEqual(t, &negZero, &posZero)

	var complexWhole big.Float
	// Parse truncates to 64 bits if precision is currently 0
	complexWhole.SetPrec(100000)
	complexWhole.Parse("12345678901234567890123456789012345678901234567890", 10)
	complexWhole.SetPrec(complexWhole.MinPrec())

	var complexMixed big.Float
	// Parse truncates to 64 bits if precision is currently 0
	complexMixed.SetPrec(100000)
	complexMixed.Parse("12345678901234567890123456789012345678901234567890"+
		".12345678901234567890123456789012345678901234567890", 10)
	complexMixed.SetPrec(complexMixed.MinPrec())

	var complexTiny big.Float
	// Parse truncates to 64 bits if precision is currently 0
	complexTiny.SetPrec(100000)
	complexTiny.Parse("0.0000000000000000000000000000000000000"+
		"12345678901234567890123456789012345678901234567890", 10)
	complexTiny.SetPrec(complexTiny.MinPrec())

	codec := internal.BigFloatCodec
	testCodecRoundTrip(t, codec, []testCase[*big.Float]{
		// example in implementation comments
		{"seven(3)", newBigFloat(7.0, 0, 3), nil},
		{"seven(4)", newBigFloat(7.0, 0, 4), nil},
		{"seven(10)", newBigFloat(7.0, 0, 10), nil},
		{"-seven(3)", newBigFloat(-7.0, 0, 3), nil},
		{"-seven(4)", newBigFloat(-7.0, 0, 4), nil},
		{"-seven(10)", newBigFloat(-7.0, 0, 10), nil},

		{"tiny", newBigFloat(12345.0, -100, 20), nil},
		{"mixed", newBigFloat(12345.0, -10, 20), nil},
		{"large", newBigFloat(12345.0, 100, 20), nil},
		{"-tiny", newBigFloat(-12345.0, -100, 20), nil},
		{"-mixed", newBigFloat(-12345.0, -10, 20), nil},
		{"-large", newBigFloat(-12345.0, 100, 20), nil},

		{"-Inf", &negInf, nil},
		{"+Inf", &posInf, nil},
		{"-0", &negZero, nil},
		{"+0", &posZero, nil},

		{"complex whole", &complexWhole, nil},
		{"complex mixed", &complexMixed, nil},
		{"complex tiny", &complexTiny, nil},
	})
}

func TestBigFloatOrdering(t *testing.T) {
	var negInf, posInf, negZero, posZero big.Float
	negInf.SetInf(true)
	posInf.SetInf(false)
	negZero.Neg(&negZero)
	assert.True(t, negZero.Signbit())
	assert.False(t, posZero.Signbit())
	assert.Equal(t, 0, negZero.Cmp(&posZero))
	assert.NotEqual(t, &negZero, &posZero)

	assert.IsIncreasing(t, [][]byte{
		encodeBigFloat(&negInf),

		// Negative Numbers
		// for the same matissa, a higher exponent (first) or precision is more negative

		// large negative numbers
		encodeBigFloat(newBigFloat(-12345.0, 10000, 21)),
		encodeBigFloat(newBigFloat(-12345.0, 10000, 20)),
		encodeBigFloat(newBigFloat(-12345.0, 10000, 19)),
		encodeBigFloat(newBigFloat(-12345.0, 9999, 21)),
		encodeBigFloat(newBigFloat(-12345.0, 9999, 20)),
		encodeBigFloat(newBigFloat(-12345.0, 9999, 19)),

		// both whole and fractional parts
		encodeBigFloat(newBigFloat(-12345.0, 10, 21)),
		encodeBigFloat(newBigFloat(-12345.0, 10, 20)),
		encodeBigFloat(newBigFloat(-12345.0, 10, 19)),

		// numbers near -7.0
		encodeBigFloat(newBigFloat(-7.1, 0, 21)),
		encodeBigFloat(newBigFloat(-7.1, 0, 20)),
		encodeBigFloat(newBigFloat(-7.0, 0, 10)), // shift 13
		encodeBigFloat(newBigFloat(-7.0, 0, 4)),  // shift 5
		encodeBigFloat(newBigFloat(-7.0, 0, 3)),  // shift 5
		encodeBigFloat(newBigFloat(-6.9, 0, 21)),
		encodeBigFloat(newBigFloat(-6.9, 0, 20)),

		// very small negative numbers
		encodeBigFloat(newBigFloat(-12345.0, -10000, 21)),
		encodeBigFloat(newBigFloat(-12345.0, -10000, 20)),
		encodeBigFloat(newBigFloat(-12345.0, -10000, 19)),

		// zeros
		encodeBigFloat(&negZero),
		encodeBigFloat(&posZero),

		// Positive Numbers
		// for the same matissa, a higher exponent (first) or precision is more positive

		// very small positive numbers
		encodeBigFloat(newBigFloat(12345.0, -10000, 19)),
		encodeBigFloat(newBigFloat(12345.0, -10000, 20)),
		encodeBigFloat(newBigFloat(12345.0, -10000, 21)),

		// numbers near 7.0
		encodeBigFloat(newBigFloat(6.9, 0, 20)),
		encodeBigFloat(newBigFloat(6.9, 0, 21)),
		encodeBigFloat(newBigFloat(7.0, 0, 3)),  // shift
		encodeBigFloat(newBigFloat(7.0, 0, 4)),  // shift 5
		encodeBigFloat(newBigFloat(7.0, 0, 10)), // shift 13
		encodeBigFloat(newBigFloat(7.1, 0, 20)),
		encodeBigFloat(newBigFloat(7.1, 0, 21)),

		// both whole and fractional parts
		encodeBigFloat(newBigFloat(12345.0, 10, 19)),
		encodeBigFloat(newBigFloat(12345.0, 10, 20)),
		encodeBigFloat(newBigFloat(12345.0, 10, 21)),

		// large positive numbers
		encodeBigFloat(newBigFloat(12345.0, 9999, 19)),
		encodeBigFloat(newBigFloat(12345.0, 9999, 20)),
		encodeBigFloat(newBigFloat(12345.0, 9999, 21)),
		encodeBigFloat(newBigFloat(12345.0, 10000, 19)),
		encodeBigFloat(newBigFloat(12345.0, 10000, 20)),
		encodeBigFloat(newBigFloat(12345.0, 10000, 21)),

		encodeBigFloat(&posInf),
	})
}
