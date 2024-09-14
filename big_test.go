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

func TestBigInt(t *testing.T) {
	t.Parallel()
	encodeSize := func(size int64) []byte {
		return lexy.Int64().Append(nil, size)
	}
	codec := lexy.BigInt()
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[*big.Int]{
		{"nil", nil, []byte{pNilFirst}},
		{"-257", big.NewInt(-257), concat([]byte{pNonNil}, encodeSize(-2),
			[]byte{0xFE, 0xFE})},
		{"-256", big.NewInt(-256), concat([]byte{pNonNil}, encodeSize(-2),
			[]byte{0xFE, 0xFF})},
		{"-255", big.NewInt(-255), concat([]byte{pNonNil}, encodeSize(-1),
			[]byte{0x00})},
		{"-254", big.NewInt(-254), concat([]byte{pNonNil}, encodeSize(-1),
			[]byte{0x01})},
		{"-2", big.NewInt(-2), concat([]byte{pNonNil}, encodeSize(-1),
			[]byte{0xFD})},
		{"-1", big.NewInt(-1), concat([]byte{pNonNil}, encodeSize(-1),
			[]byte{0xFE})},
		{"0", big.NewInt(0), concat([]byte{pNonNil}, encodeSize(0),
			[]byte{})},
		{"+1", big.NewInt(1), concat([]byte{pNonNil}, encodeSize(1),
			[]byte{0x01})},
		{"+2", big.NewInt(2), concat([]byte{pNonNil}, encodeSize(1),
			[]byte{0x02})},
		{"254", big.NewInt(254), concat([]byte{pNonNil}, encodeSize(1),
			[]byte{0xFE})},
		{"255", big.NewInt(255), concat([]byte{pNonNil}, encodeSize(1),
			[]byte{0xFF})},
		{"256", big.NewInt(256), concat([]byte{pNonNil}, encodeSize(2),
			[]byte{0x01, 0x00})},
		{"257", big.NewInt(257), concat([]byte{pNonNil}, encodeSize(2),
			[]byte{0x01, 0x01})},
	})

	testCodec(t, codec, fillTestData(codec, []testCase[*big.Int]{
		{"big positive", newBigInt(manyDigits), nil},
		{"big negative", newBigInt("-" + manyDigits), nil},
	}))
}

func TestBigIntOrdering(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.BigInt(), []testCase[*big.Int]{
		{"nil", nil, nil},
		{"-12345", newBigInt("-12345"), nil},
		{"-12344", newBigInt("-12344"), nil},
		{"-12343", newBigInt("-12343"), nil},
		{"-257", newBigInt("-257"), nil},
		{"-256", newBigInt("-256"), nil},
		{"-255", newBigInt("-255"), nil},
		{"-1", newBigInt("-1"), nil},
		{"0", newBigInt("0"), nil},
		{"1", newBigInt("1"), nil},
		{"255", newBigInt("255"), nil},
		{"256", newBigInt("256"), nil},
		{"257", newBigInt("257"), nil},
		{"12343", newBigInt("12343"), nil},
		{"12344", newBigInt("12344"), nil},
		{"12345", newBigInt("12345"), nil},
	})
}

func TestBigIntNilsLast(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.NilsLast(lexy.BigInt()), []testCase[*big.Int]{
		{"-12345", newBigInt("-12345"), nil},
		{"0", newBigInt("0"), nil},
		{"12345", newBigInt("12345"), nil},
		{"nil", nil, nil},
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
	//nolint:dogsled,errcheck
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
	assert.False(t, codec.RequiresTerminator())
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

	// For the same matissa, a higher exponent (first) or precision is closer to infinity.
	testOrdering(t, lexy.BigFloat(), []testCase[*big.Float]{
		{"nil", nil, nil},
		{"-Inf", &negInf, nil},

		// large negative numbers
		{"-12345.0 * 2^10000 (21)", newBigFloat64(-12345.0, 10000, 21), nil},
		{"-12345.0 * 2^10000 (20)", newBigFloat64(-12345.0, 10000, 20), nil},
		{"-12345.0 * 2^10000 (19)", newBigFloat64(-12345.0, 10000, 19), nil},
		{"-12345.0 * 2^9999 (21)", newBigFloat64(-12345.0, 9999, 21), nil},
		{"-12345.0 * 2^9999 (20)", newBigFloat64(-12345.0, 9999, 20), nil},
		{"-12345.0 * 2^9999 (19)", newBigFloat64(-12345.0, 9999, 19), nil},

		// both whole and fractional parts
		{"-12345.0 * 2^10 (21)", newBigFloat64(-12345.0, 10, 21), nil},
		{"-12345.0 * 2^10 (20)", newBigFloat64(-12345.0, 10, 20), nil},
		{"-12345.0 * 2^10 (19)", newBigFloat64(-12345.0, 10, 19), nil},

		// numbers near -7.0
		{"-7.1 * 2^0 (21)", newBigFloat64(-7.1, 0, 21), nil},
		{"-7.1 * 2^0 (20)", newBigFloat64(-7.1, 0, 20), nil},
		{"-7.0 * 2^0 (10)", newBigFloat64(-7.0, 0, 10), nil}, // shift 13
		{"-7.0 * 2^0 (4)", newBigFloat64(-7.0, 0, 4), nil},   // shift 5
		{"-7.0 * 2^0 (3)", newBigFloat64(-7.0, 0, 3), nil},   // shift 5
		{"-6.9 * 2^0 (21)", newBigFloat64(-6.9, 0, 21), nil},
		{"-6.9 * 2^0 (20)", newBigFloat64(-6.9, 0, 20), nil},

		// very small negative numbers
		{"-12345.0 * 2^-10000 (21)", newBigFloat64(-12345.0, -10000, 21), nil},
		{"-12345.0 * 2^-10000 (20)", newBigFloat64(-12345.0, -10000, 20), nil},
		{"-12345.0 * 2^-10000 (19)", newBigFloat64(-12345.0, -10000, 19), nil},

		// zeros
		{"-0.0", &negZero, nil},
		{"+0.0", &posZero, nil},

		// very small positive numbers
		{"12345.0 * 2^-10000 (19)", newBigFloat64(12345.0, -10000, 19), nil},
		{"12345.0 * 2^-10000 (20)", newBigFloat64(12345.0, -10000, 20), nil},
		{"12345.0 * 2^-10000 (21)", newBigFloat64(12345.0, -10000, 21), nil},

		// numbers near 7.0
		{"6.9 * 2^0 (20)", newBigFloat64(6.9, 0, 20), nil},
		{"6.9 * 2^0 (21)", newBigFloat64(6.9, 0, 21), nil},
		{"7.0 * 2^0 (3)", newBigFloat64(7.0, 0, 3), nil},   // shift
		{"7.0 * 2^0 (4)", newBigFloat64(7.0, 0, 4), nil},   // shift 5
		{"7.0 * 2^0 (10)", newBigFloat64(7.0, 0, 10), nil}, // shift 13
		{"7.1 * 2^0 (20)", newBigFloat64(7.1, 0, 20), nil},
		{"7.1 * 2^0 (21)", newBigFloat64(7.1, 0, 21), nil},

		// both whole and fractional parts
		{"12345.0 * 2^10 (19)", newBigFloat64(12345.0, 10, 19), nil},
		{"12345.0 * 2^10 (20)", newBigFloat64(12345.0, 10, 20), nil},
		{"12345.0 * 2^10 (21)", newBigFloat64(12345.0, 10, 21), nil},

		// large positive numbers
		{"12345.0 * 2^9999 (19)", newBigFloat64(12345.0, 9999, 19), nil},
		{"12345.0 * 2^9999 (20)", newBigFloat64(12345.0, 9999, 20), nil},
		{"12345.0 * 2^9999 (21)", newBigFloat64(12345.0, 9999, 21), nil},
		{"12345.0 * 2^10000 (19)", newBigFloat64(12345.0, 10000, 19), nil},
		{"12345.0 * 2^10000 (20)", newBigFloat64(12345.0, 10000, 20), nil},
		{"12345.0 * 2^10000 (21)", newBigFloat64(12345.0, 10000, 21), nil},

		{"+Inf", &posInf, nil},
	})
}

func TestBigFloatNilsLast(t *testing.T) {
	t.Parallel()
	var negInf, posInf, posZero big.Float
	negInf.SetInf(true)
	posInf.SetInf(false)
	testOrdering(t, lexy.NilsLast(lexy.BigFloat()), []testCase[*big.Float]{
		{"-Inf", &negInf, nil},
		{"+0.0", &posZero, nil},
		{"+Inf", &posInf, nil},
		{"nil", nil, nil},
	})
}

func newBigRat(num, denom string) *big.Rat {
	var value big.Rat
	return value.SetFrac(newBigInt(num), newBigInt(denom))
}

func TestBigRat(t *testing.T) {
	t.Parallel()
	codec := lexy.BigRat()
	assert.False(t, codec.RequiresTerminator())
	// Note that big.Rat normalizes values when set using SetFrac.
	// So 2/4 => 1/2, and 0/100 => 0/1
	testCodec(t, codec, fillTestData(codec, []testCase[*big.Rat]{
		{"-1/3", newBigRat("-1", "3"), nil},
		{"0/123", newBigRat("0", "123"), nil},
		{"5432/42", newBigRat("5432", "42"), nil},
	}))
}

func TestBigRatOrdering(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.BigRat(), []testCase[*big.Rat]{
		{"nil", nil, nil},
		{"-1/1", newBigRat("-1", "1"), nil},
		{"-1/2", newBigRat("-1", "2"), nil},
		{"0/1", newBigRat("0", "1"), nil},
		{"1/1", newBigRat("1", "1"), nil},
		{"1/2", newBigRat("1", "2"), nil},
	})
}

func TestBigRatNilsLast(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.NilsLast(lexy.BigRat()), []testCase[*big.Rat]{
		{"-1/1", newBigRat("-1", "1"), nil},
		{"0/1", newBigRat("0", "1"), nil},
		{"1/2", newBigRat("1", "2"), nil},
		{"nil", nil, nil},
	})
}
