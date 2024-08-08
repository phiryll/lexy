package lexy_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Seed values for different types.
var (
	seedsUint8  = []uint8{0, 1, math.MaxUint8}
	seedsUint16 = []uint16{0, 1, math.MaxUint16}
	seedsUint32 = []uint32{0, 1, math.MaxUint32}
	seedsUint64 = []uint64{0, 1, math.MaxUint64}
	seedsInt8   = []int8{0, 1, -1, math.MinInt8, math.MaxInt8}
	seedsInt16  = []int16{0, 1, -1, math.MinInt16, math.MaxInt16}
	seedsInt32  = []int32{0, 1, -1, math.MinInt32, math.MaxInt32}
	seedsInt64  = []int64{0, 1, -1, math.MinInt64, math.MaxInt64}

	// Fuzzing bit patterns instead of floats
	// because go's float fuzzer only generates one pattern for NaN.
	seedsFloat32 = []uint32{
		math.Float32bits(math.MaxFloat32),
		math.Float32bits(math.SmallestNonzeroFloat32),
		math.Float32bits(float32(math.Inf(1))),
		math.Float32bits(float32(math.NaN())),
		math.Float32bits(0.0),
		math.Float32bits(123.456e+23),
		math.Float32bits(-math.MaxFloat32),
		math.Float32bits(-math.SmallestNonzeroFloat32),
		math.Float32bits(float32(math.Inf(-1))),
		math.Float32bits(-float32(math.NaN())),
		math.Float32bits(float32(math.Copysign(0.0, -1.0))),
		math.Float32bits(-123.456e+23),
	}
	seedsFloat64 = []uint64{
		math.Float64bits(math.MaxFloat64),
		math.Float64bits(math.SmallestNonzeroFloat64),
		math.Float64bits(math.Inf(1)),
		math.Float64bits(math.NaN()),
		math.Float64bits(0.0),
		math.Float64bits(123.456e+23),
		math.Float64bits(-math.MaxFloat64),
		math.Float64bits(-math.SmallestNonzeroFloat64),
		math.Float64bits(math.Inf(-1)),
		math.Float64bits(-math.NaN()),
		math.Float64bits(math.Copysign(0.0, -1.0)),
		math.Float64bits(-123.456e+23),
	}

	seedsString = []string{
		"",
		"q",
		"\xFE",
		"\x00",
		"\x01",
		"\xFF",
		"a b c",
		"a b d",
		"a/\xFF34\x009``[*\x01#)2f\xFEmn",
	}

	seedsBytes = [][]byte{
		nil,
		{},
		{0},
		{1},
		{254},
		{255},
		{254, 0, 34, 72, 0, 1, 0, 255, 0, 17},
	}
)

// Helper function somewhat duplicating cmp.Compare (go 1.21, so trying to avoid)
func compare[T uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64 | string](x, y T) int {
	switch {
	case x < y:
		return -1
	case x == y:
		return 0
	default:
		return 1
	}
}

// translates representations, used for bits<->float
type converter[T, U any] interface {
	To(t T) U
	From(u U) T
	Cmp(a, b T) int
}

var (
	f32Conv   converter[float32, uint32] = float32Converter{}
	f64Conv   converter[float64, uint64] = float64Converter{}
	neg32Conv converter[float32, uint32] = negFloat32Converter{}
)

type float32Converter struct{}

func (c float32Converter) To(f float32) uint32   { return math.Float32bits(f) }
func (c float32Converter) From(u uint32) float32 { return math.Float32frombits(u) }
func (c float32Converter) Cmp(a, b float32) int {
	return cmpFloats(math.Float32bits, a, b)
}

type float64Converter struct{}

func (c float64Converter) To(f float64) uint64   { return math.Float64bits(f) }
func (c float64Converter) From(u uint64) float64 { return math.Float64frombits(u) }
func (c float64Converter) Cmp(a, b float64) int {
	return cmpFloats(math.Float64bits, a, b)
}

type negFloat32Converter struct{}

func (c negFloat32Converter) To(f float32) uint32 { return f32Conv.To(negativeFloat32(f)) }
func (c negFloat32Converter) From(u uint32) float32 {
	return negativeFloat32(f32Conv.From(u))
}
func (c negFloat32Converter) Cmp(a, b float32) int {
	return f32Conv.Cmp(b, a)
}

// Functions to add seed values to the fuzzer.

func addValues[T any](f *testing.F, values ...T) {
	for _, x := range values {
		f.Add(x)
	}
}

// used for testing cmp(v1, v2)
func addUnorderedPairs[T any](f *testing.F, values ...T) {
	for i, x := range values {
		for _, y := range values[i+1:] {
			f.Add(x, y)
		}
	}
}

// These fuzzers test the encode-decode round trip.

func valueTesterFor[T any](codec lexy.Codec[T]) func(*testing.T, T) {
	return func(t *testing.T, value T) {
		b, err := lexy.Encode(codec, value)
		require.NoError(t, err)
		got, err := lexy.Decode(codec, b)
		require.NoError(t, err)
		assert.IsType(t, value, got)
		assert.Equal(t, value, got)
	}
}

// Implements ordering semantics of the float Codecs, mostly without encoding them.
func cmpFloats[T float32 | float64, U uint32 | uint64](toBits func(T) U, a, b T) int {
	aBits := toBits(a)
	bBits := toBits(b)
	if aBits == bBits {
		return 0
	}
	aSign := math.Signbit(float64(a)) // true if negative or -0.0
	if aSign != math.Signbit(float64(b)) {
		if aSign {
			return -1
		}
		return 1
	}
	// at this point, a != b and they have the same sign, only special case is NaN
	switch {
	case math.IsNaN(float64(a)) && math.IsNaN(float64(b)):
		if aSign {
			// Codec flips all bits, compare in reverse order
			return compare(bBits, aBits)
		} else {
			// Codec flips the high bit, compare as signed ints
			return compare(int64(aBits), int64(bBits))
		}
	case math.IsNaN(float64(a)):
		if aSign {
			return -1
		}
		return 1
	case math.IsNaN(float64(b)):
		if aSign {
			return 1
		}
		return -1
	default:
		if a < b {
			return -1
		}
		return 1
	}
}

func valueTesterForConv[T, U any](codec lexy.Codec[T], conv converter[T, U]) func(*testing.T, U) {
	return func(t *testing.T, repr U) {
		value := conv.From(repr)
		b, err := lexy.Encode(codec, value)
		require.NoError(t, err)
		got, err := lexy.Decode(codec, b)
		require.NoError(t, err)
		assert.IsType(t, value, got)
		assert.Equal(t, conv.To(value), conv.To(got), "values not equal: %#v, %#v", value, got)
	}
}

func FuzzUint8(f *testing.F) {
	addValues(f, seedsUint8...)
	f.Fuzz(valueTesterFor(lexy.MakeUint8[uint8]()))
}

func FuzzUint16(f *testing.F) {
	addValues(f, seedsUint16...)
	f.Fuzz(valueTesterFor(lexy.MakeUint16[uint16]()))
}

func FuzzUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(valueTesterFor(lexy.MakeUint32[uint32]()))
}

func FuzzUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(valueTesterFor(lexy.MakeUint64[uint64]()))
}

func FuzzInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(valueTesterFor(lexy.MakeInt8[int8]()))
}

func FuzzInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(valueTesterFor(lexy.MakeInt16[int16]()))
}

func FuzzInt32(f *testing.F) {
	addValues(f, seedsInt32...)
	f.Fuzz(valueTesterFor(lexy.MakeInt32[int32]()))
}

func FuzzInt64(f *testing.F) {
	addValues(f, seedsInt64...)
	f.Fuzz(valueTesterFor(lexy.MakeInt64[int64]()))
}

func FuzzFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(valueTesterForConv(lexy.MakeFloat32[float32](), f32Conv))
}

func FuzzFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(valueTesterForConv(lexy.MakeFloat64[float64](), f64Conv))
}

func FuzzString(f *testing.F) {
	addValues(f, seedsString...)
	f.Fuzz(valueTesterFor(lexy.MakeString[string]()))
}

func FuzzBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.MakeBytes[[]byte]()))
}

func FuzzNegUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.MakeUint32[uint32]())))
}

func FuzzNegInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.MakeInt8[int8]())))
}

func FuzzNegFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(valueTesterForConv(lexy.Negate(lexy.MakeFloat64[float64]()), f64Conv))
}

func FuzzNegBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.MakeBytes[[]byte]())))
}

func FuzzTerminateUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.MakeUint64[uint64]())))
}

func FuzzTerminateInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.MakeInt16[int16]())))
}

func FuzzTerminateFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(valueTesterForConv(lexy.Terminate(lexy.MakeFloat32[float32]()), f32Conv))
}

func FuzzTerminateBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.MakeBytes[[]byte]())))
}

// These fuzzers test that the encoding order is consistent with the value order.

func pairTesterFor[T any](codec lexy.Codec[T], cmp func(T, T) int) func(*testing.T, T, T) {
	return func(t *testing.T, a, b T) {
		aEncoded, err := lexy.Encode(codec, a)
		require.NoError(t, err)
		bEncoded, err := lexy.Encode(codec, b)
		require.NoError(t, err)
		assert.Equal(t, cmp(a, b), bytes.Compare(aEncoded, bEncoded),
			"values not comparing correctly: %#v(%x), %#v(%x)", a, aEncoded, b, bEncoded)
	}
}

func pairTesterForConv[T, U any](codec lexy.Codec[T], conv converter[T, U]) func(*testing.T, U, U) {
	f := pairTesterFor(codec, conv.Cmp)
	return func(t *testing.T, a, b U) {
		f(t, conv.From(a), conv.From(b))
	}
}

// because bytes.Compare(nil, {}) == 0
func cmpBytes(a, b []byte) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	default:
		return bytes.Compare(a, b)
	}
}

func FuzzCmpUint8(f *testing.F) {
	addUnorderedPairs(f, seedsUint8...)
	f.Fuzz(pairTesterFor(lexy.MakeUint8[uint8](), compare[uint8]))
}

func FuzzCmpUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(pairTesterFor(lexy.MakeUint16[uint16](), compare[uint16]))
}

func FuzzCmpUint32(f *testing.F) {
	addUnorderedPairs(f, seedsUint32...)
	f.Fuzz(pairTesterFor(lexy.MakeUint32[uint32](), compare[uint32]))
}

func FuzzCmpUint64(f *testing.F) {
	addUnorderedPairs(f, seedsUint64...)
	f.Fuzz(pairTesterFor(lexy.MakeUint64[uint64](), compare[uint64]))
}

func FuzzCmpInt8(f *testing.F) {
	addUnorderedPairs(f, seedsInt8...)
	f.Fuzz(pairTesterFor(lexy.MakeInt8[int8](), compare[int8]))
}

func FuzzCmpInt16(f *testing.F) {
	addUnorderedPairs(f, seedsInt16...)
	f.Fuzz(pairTesterFor(lexy.MakeInt16[int16](), compare[int16]))
}

func FuzzCmpInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(pairTesterFor(lexy.MakeInt32[int32](), compare[int32]))
}

func FuzzCmpInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(pairTesterFor(lexy.MakeInt64[int64](), compare[int64]))
}

func FuzzCmpFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(pairTesterForConv(lexy.MakeFloat32[float32](), f32Conv))
}

func FuzzCmpFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(pairTesterForConv(lexy.MakeFloat64[float64](), f64Conv))
}

func FuzzCmpString(f *testing.F) {
	addUnorderedPairs(f, seedsString...)
	f.Fuzz(pairTesterFor(lexy.MakeString[string](), compare[string]))
}

func FuzzCmpBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.MakeBytes[[]byte](), cmpBytes))
}

func negCmp[T any](cmp func(T, T) int) func(T, T) int {
	return func(a, b T) int {
		return cmp(b, a)
	}
}

func negativeFloat32(f float32) float32 {
	f64 := float64(f)
	if math.Signbit(f64) {
		return float32(math.Copysign(f64, 1.0))
	}
	return float32(math.Copysign(f64, -1.0))
}

func FuzzCmpNegUint8(f *testing.F) {
	addUnorderedPairs(f, seedsUint8...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.MakeUint8[uint8]()), negCmp(compare[uint8])))
}

func FuzzCmpNegInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.MakeInt32[int32]()), negCmp(compare[int32])))
}

func FuzzCmpNegFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(pairTesterForConv(lexy.Negate(lexy.MakeFloat32[float32]()), neg32Conv))
}

func FuzzCmpNegBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.MakeBytes[[]byte]()), negCmp(cmpBytes)))
}

func FuzzCmpTerminateUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.MakeUint16[uint16]()), compare[uint16]))
}

func FuzzCmpTerminateInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.MakeInt64[int64]()), compare[int64]))
}

func FuzzCmpTerminateFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(pairTesterForConv(lexy.Terminate(lexy.MakeFloat64[float64]()), f64Conv))
}

func FuzzCmpTerminateBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.MakeBytes[[]byte]()), cmpBytes))
}
