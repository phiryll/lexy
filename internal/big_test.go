package internal_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func newBigInt(s string) *big.Int {
	var value big.Int
	value.SetString(s, 10)
	return &value
}

func TestBigInt(t *testing.T) {
	codec := internal.BigIntCodec(true)
	encodeSize := encoderFor(int64Codec)
	prefix := []byte{pNonNil}
	testCodec(t, codec, []testCase[*big.Int]{
		{"nil", nil, []byte{pNilFirst}},
		{"-257", big.NewInt(-257), slices.Concat(prefix, encodeSize(-2),
			[]byte{0xFE, 0xFE})},
		{"-256", big.NewInt(-256), slices.Concat(prefix, encodeSize(-2),
			[]byte{0xFE, 0xFF})},
		{"-255", big.NewInt(-255), slices.Concat(prefix, encodeSize(-1),
			[]byte{0x00})},
		{"-254", big.NewInt(-254), slices.Concat(prefix, encodeSize(-1),
			[]byte{0x01})},
		{"-2", big.NewInt(-2), slices.Concat(prefix, encodeSize(-1),
			[]byte{0xFD})},
		{"-1", big.NewInt(-1), slices.Concat(prefix, encodeSize(-1),
			[]byte{0xFE})},
		{"0", big.NewInt(0), slices.Concat(prefix, encodeSize(0),
			[]byte{})},
		{"+1", big.NewInt(1), slices.Concat(prefix, encodeSize(1),
			[]byte{0x01})},
		{"+2", big.NewInt(2), slices.Concat(prefix, encodeSize(1),
			[]byte{0x02})},
		{"254", big.NewInt(254), slices.Concat(prefix, encodeSize(1),
			[]byte{0xFE})},
		{"255", big.NewInt(255), slices.Concat(prefix, encodeSize(1),
			[]byte{0xFF})},
		{"256", big.NewInt(256), slices.Concat(prefix, encodeSize(2),
			[]byte{0x01, 0x00})},
		{"257", big.NewInt(257), slices.Concat(prefix, encodeSize(2),
			[]byte{0x01, 0x01})},
	})

	testCodecRoundTrip(t, codec, []testCase[*big.Int]{
		{"big positive", newBigInt("1234567890123456789012345678901234567890"), nil},
		{"big negative", newBigInt("-1234567890123456789012345678901234567890"), nil},
	})
}

func TestBigIntOrdering(t *testing.T) {
	encode := encoderFor(internal.BigIntCodec(true))
	assert.IsIncreasing(t, [][]byte{
		encode(nil),
		encode(newBigInt("-12345")),
		encode(newBigInt("-12344")),
		encode(newBigInt("-12343")),
		encode(newBigInt("-257")),
		encode(newBigInt("-256")),
		encode(newBigInt("-255")),
		encode(newBigInt("-1")),
		encode(newBigInt("0")),
		encode(newBigInt("1")),
		encode(newBigInt("255")),
		encode(newBigInt("256")),
		encode(newBigInt("257")),
		encode(newBigInt("12343")),
		encode(newBigInt("12344")),
		encode(newBigInt("12345")),
	})
}

func TestBigIntNilsLast(t *testing.T) {
	encodeFirst := encoderFor(internal.BigIntCodec(true))
	encodeLast := encoderFor(internal.BigIntCodec(false))
	assert.IsIncreasing(t, [][]byte{
		encodeFirst(nil),
		encodeFirst(newBigInt("-12345")),
		encodeFirst(newBigInt("0")),
		encodeFirst(newBigInt("12345")),
	})
	assert.IsIncreasing(t, [][]byte{
		encodeLast(newBigInt("-12345")),
		encodeLast(newBigInt("0")),
		encodeLast(newBigInt("12345")),
		encodeLast(nil),
	})
}

func newBigFloat(f float64, shift int, prec uint) *big.Float {
	value := big.NewFloat(f)
	value.SetPrec(prec)
	value.SetMantExp(value, shift)
	return value
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

	codec := internal.BigFloatCodec(true)
	testCodecRoundTrip(t, codec, []testCase[*big.Float]{
		{"nil", nil, nil},
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

	encode := encoderFor(internal.BigFloatCodec(true))
	assert.IsIncreasing(t, [][]byte{
		encode(nil),
		encode(&negInf),

		// Negative Numbers
		// for the same matissa, a higher exponent (first) or precision is more negative

		// large negative numbers
		encode(newBigFloat(-12345.0, 10000, 21)),
		encode(newBigFloat(-12345.0, 10000, 20)),
		encode(newBigFloat(-12345.0, 10000, 19)),
		encode(newBigFloat(-12345.0, 9999, 21)),
		encode(newBigFloat(-12345.0, 9999, 20)),
		encode(newBigFloat(-12345.0, 9999, 19)),

		// both whole and fractional parts
		encode(newBigFloat(-12345.0, 10, 21)),
		encode(newBigFloat(-12345.0, 10, 20)),
		encode(newBigFloat(-12345.0, 10, 19)),

		// numbers near -7.0
		encode(newBigFloat(-7.1, 0, 21)),
		encode(newBigFloat(-7.1, 0, 20)),
		encode(newBigFloat(-7.0, 0, 10)), // shift 13
		encode(newBigFloat(-7.0, 0, 4)),  // shift 5
		encode(newBigFloat(-7.0, 0, 3)),  // shift 5
		encode(newBigFloat(-6.9, 0, 21)),
		encode(newBigFloat(-6.9, 0, 20)),

		// very small negative numbers
		encode(newBigFloat(-12345.0, -10000, 21)),
		encode(newBigFloat(-12345.0, -10000, 20)),
		encode(newBigFloat(-12345.0, -10000, 19)),

		// zeros
		encode(&negZero),
		encode(&posZero),

		// Positive Numbers
		// for the same matissa, a higher exponent (first) or precision is more positive

		// very small positive numbers
		encode(newBigFloat(12345.0, -10000, 19)),
		encode(newBigFloat(12345.0, -10000, 20)),
		encode(newBigFloat(12345.0, -10000, 21)),

		// numbers near 7.0
		encode(newBigFloat(6.9, 0, 20)),
		encode(newBigFloat(6.9, 0, 21)),
		encode(newBigFloat(7.0, 0, 3)),  // shift
		encode(newBigFloat(7.0, 0, 4)),  // shift 5
		encode(newBigFloat(7.0, 0, 10)), // shift 13
		encode(newBigFloat(7.1, 0, 20)),
		encode(newBigFloat(7.1, 0, 21)),

		// both whole and fractional parts
		encode(newBigFloat(12345.0, 10, 19)),
		encode(newBigFloat(12345.0, 10, 20)),
		encode(newBigFloat(12345.0, 10, 21)),

		// large positive numbers
		encode(newBigFloat(12345.0, 9999, 19)),
		encode(newBigFloat(12345.0, 9999, 20)),
		encode(newBigFloat(12345.0, 9999, 21)),
		encode(newBigFloat(12345.0, 10000, 19)),
		encode(newBigFloat(12345.0, 10000, 20)),
		encode(newBigFloat(12345.0, 10000, 21)),

		encode(&posInf),
	})
}

func newBigRat(num, denom string) *big.Rat {
	var value big.Rat
	return value.SetFrac(newBigInt(num), newBigInt(denom))
}

func TestBigRat(t *testing.T) {
	// Note that big.Rat normalizes values when set using SetFrac.
	// So 2/4 => 1/2, and 0/100 => 0/1
	codec := internal.BigRatCodec(true)
	testCodecRoundTrip(t, codec, []testCase[*big.Rat]{
		{"-1/3", newBigRat("-1", "3"), nil},
		{"0/123", newBigRat("0", "123"), nil},
		{"5432/42", newBigRat("5432", "42"), nil},
	})

	encode := encoderFor(codec)
	assert.IsIncreasing(t, [][]byte{
		encode(nil),
		encode(newBigRat("-1", "1")),
		encode(newBigRat("-1", "2")),
		encode(newBigRat("0", "1")),
		encode(newBigRat("1", "1")),
		encode(newBigRat("1", "2")),
	})
}
