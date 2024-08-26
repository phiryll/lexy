package lexy_test

import (
	"math/big"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

const (
	manyZeros  = "00000000000000000000000000000000000000000000000000"
	manyDigits = "12345678901234567890123456789012345678901234567890"
)

func newBigInt(s string) *big.Int {
	var value big.Int
	value.SetString(s, 10)
	return &value
}

func concatNonNil(slices ...[]byte) []byte {
	result := []byte{pNonNil}
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func TestBigInt(t *testing.T) {
	t.Parallel()
	codec := lexy.BigInt()
	encodeSize := encoderFor(lexy.Int64())
	testCodec(t, codec, []testCase[*big.Int]{
		{"nil", nil, []byte{pNilFirst}},
		{"-257", big.NewInt(-257), concatNonNil(encodeSize(-2),
			[]byte{0xFE, 0xFE})},
		{"-256", big.NewInt(-256), concatNonNil(encodeSize(-2),
			[]byte{0xFE, 0xFF})},
		{"-255", big.NewInt(-255), concatNonNil(encodeSize(-1),
			[]byte{0x00})},
		{"-254", big.NewInt(-254), concatNonNil(encodeSize(-1),
			[]byte{0x01})},
		{"-2", big.NewInt(-2), concatNonNil(encodeSize(-1),
			[]byte{0xFD})},
		{"-1", big.NewInt(-1), concatNonNil(encodeSize(-1),
			[]byte{0xFE})},
		{"0", big.NewInt(0), concatNonNil(encodeSize(0),
			[]byte{})},
		{"+1", big.NewInt(1), concatNonNil(encodeSize(1),
			[]byte{0x01})},
		{"+2", big.NewInt(2), concatNonNil(encodeSize(1),
			[]byte{0x02})},
		{"254", big.NewInt(254), concatNonNil(encodeSize(1),
			[]byte{0xFE})},
		{"255", big.NewInt(255), concatNonNil(encodeSize(1),
			[]byte{0xFF})},
		{"256", big.NewInt(256), concatNonNil(encodeSize(2),
			[]byte{0x01, 0x00})},
		{"257", big.NewInt(257), concatNonNil(encodeSize(2),
			[]byte{0x01, 0x01})},
	})

	testCodec(t, codec, fillTestData(codec, []testCase[*big.Int]{
		{"big positive", newBigInt(manyDigits), nil},
		{"big negative", newBigInt("-" + manyDigits), nil},
	}))
}

func TestBigIntOrdering(t *testing.T) {
	t.Parallel()
	encode := encoderFor(lexy.BigInt())
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
	t.Parallel()
	encodeFirst := encoderFor(lexy.BigInt())
	encodeLast := encoderFor(lexy.NilsLast(lexy.BigInt()))
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

func newBigFloat64(f float64, shift int, prec uint) *big.Float {
	value := big.NewFloat(f)
	value.SetPrec(prec)
	value.SetMantExp(value, shift)
	return value
}

func newBigFloat(s string) *big.Float {
	var value big.Float
	// Parse truncates to 64 bits if precision is currently 0.
	value.SetPrec(100000)
	//nolint:dogsled,errcheck,gosec
	_, _, _ = value.Parse(s, 10)
	value.SetPrec(value.MinPrec())
	return &value
}

func TestBigFloat(t *testing.T) {
	t.Parallel()
	var negInf, posInf, negZero, posZero big.Float
	negInf.SetInf(true)
	posInf.SetInf(false)
	negZero.Neg(&negZero)
	assert.True(t, negZero.Signbit())
	assert.False(t, posZero.Signbit())
	assert.Equal(t, 0, negZero.Cmp(&posZero))
	assert.NotEqual(t, &negZero, &posZero)

	wholeNumber := newBigFloat(manyDigits + manyDigits)
	mixedNumber := newBigFloat(manyDigits + "." + manyDigits)
	smallNumber := newBigFloat("0." + manyZeros + manyDigits)

	codec := lexy.BigFloat()
	testCodec(t, codec, fillTestData(codec, []testCase[*big.Float]{
		{"nil", nil, nil},
		// example in implementation comments
		{"seven(3)", newBigFloat64(7.0, 0, 3), nil},
		{"seven(4)", newBigFloat64(7.0, 0, 4), nil},
		{"seven(10)", newBigFloat64(7.0, 0, 10), nil},
		{"-seven(3)", newBigFloat64(-7.0, 0, 3), nil},
		{"-seven(4)", newBigFloat64(-7.0, 0, 4), nil},
		{"-seven(10)", newBigFloat64(-7.0, 0, 10), nil},

		{"tiny", newBigFloat64(12345.0, -100, 20), nil},
		{"mixed", newBigFloat64(12345.0, -10, 20), nil},
		{"large", newBigFloat64(12345.0, 100, 20), nil},
		{"-tiny", newBigFloat64(-12345.0, -100, 20), nil},
		{"-mixed", newBigFloat64(-12345.0, -10, 20), nil},
		{"-large", newBigFloat64(-12345.0, 100, 20), nil},

		{"-Inf", &negInf, nil},
		{"+Inf", &posInf, nil},
		{"-0", &negZero, nil},
		{"+0", &posZero, nil},

		{"long whole", wholeNumber, nil},
		{"long mixed", mixedNumber, nil},
		{"long small", smallNumber, nil},
	}))
}

//nolint:funlen
func TestBigFloatOrdering(t *testing.T) {
	t.Parallel()
	var negInf, posInf, negZero, posZero big.Float
	negInf.SetInf(true)
	posInf.SetInf(false)
	negZero.Neg(&negZero)
	assert.True(t, negZero.Signbit())
	assert.False(t, posZero.Signbit())
	assert.Equal(t, 0, negZero.Cmp(&posZero))
	assert.NotEqual(t, &negZero, &posZero)

	encode := encoderFor(lexy.BigFloat())
	assert.IsIncreasing(t, [][]byte{
		encode(nil),
		encode(&negInf),

		// Negative Numbers
		// for the same matissa, a higher exponent (first) or precision is more negative

		// large negative numbers
		encode(newBigFloat64(-12345.0, 10000, 21)),
		encode(newBigFloat64(-12345.0, 10000, 20)),
		encode(newBigFloat64(-12345.0, 10000, 19)),
		encode(newBigFloat64(-12345.0, 9999, 21)),
		encode(newBigFloat64(-12345.0, 9999, 20)),
		encode(newBigFloat64(-12345.0, 9999, 19)),

		// both whole and fractional parts
		encode(newBigFloat64(-12345.0, 10, 21)),
		encode(newBigFloat64(-12345.0, 10, 20)),
		encode(newBigFloat64(-12345.0, 10, 19)),

		// numbers near -7.0
		encode(newBigFloat64(-7.1, 0, 21)),
		encode(newBigFloat64(-7.1, 0, 20)),
		encode(newBigFloat64(-7.0, 0, 10)), // shift 13
		encode(newBigFloat64(-7.0, 0, 4)),  // shift 5
		encode(newBigFloat64(-7.0, 0, 3)),  // shift 5
		encode(newBigFloat64(-6.9, 0, 21)),
		encode(newBigFloat64(-6.9, 0, 20)),

		// very small negative numbers
		encode(newBigFloat64(-12345.0, -10000, 21)),
		encode(newBigFloat64(-12345.0, -10000, 20)),
		encode(newBigFloat64(-12345.0, -10000, 19)),

		// zeros
		encode(&negZero),
		encode(&posZero),

		// Positive Numbers
		// for the same matissa, a higher exponent (first) or precision is more positive

		// very small positive numbers
		encode(newBigFloat64(12345.0, -10000, 19)),
		encode(newBigFloat64(12345.0, -10000, 20)),
		encode(newBigFloat64(12345.0, -10000, 21)),

		// numbers near 7.0
		encode(newBigFloat64(6.9, 0, 20)),
		encode(newBigFloat64(6.9, 0, 21)),
		encode(newBigFloat64(7.0, 0, 3)),  // shift
		encode(newBigFloat64(7.0, 0, 4)),  // shift 5
		encode(newBigFloat64(7.0, 0, 10)), // shift 13
		encode(newBigFloat64(7.1, 0, 20)),
		encode(newBigFloat64(7.1, 0, 21)),

		// both whole and fractional parts
		encode(newBigFloat64(12345.0, 10, 19)),
		encode(newBigFloat64(12345.0, 10, 20)),
		encode(newBigFloat64(12345.0, 10, 21)),

		// large positive numbers
		encode(newBigFloat64(12345.0, 9999, 19)),
		encode(newBigFloat64(12345.0, 9999, 20)),
		encode(newBigFloat64(12345.0, 9999, 21)),
		encode(newBigFloat64(12345.0, 10000, 19)),
		encode(newBigFloat64(12345.0, 10000, 20)),
		encode(newBigFloat64(12345.0, 10000, 21)),

		encode(&posInf),
	})
}

func newBigRat(num, denom string) *big.Rat {
	var value big.Rat
	return value.SetFrac(newBigInt(num), newBigInt(denom))
}

func TestBigRat(t *testing.T) {
	t.Parallel()
	codec := lexy.BigRat()
	// Note that big.Rat normalizes values when set using SetFrac.
	// So 2/4 => 1/2, and 0/100 => 0/1
	testCodec(t, codec, fillTestData(codec, []testCase[*big.Rat]{
		{"-1/3", newBigRat("-1", "3"), nil},
		{"0/123", newBigRat("0", "123"), nil},
		{"5432/42", newBigRat("5432", "42"), nil},
	}))

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
