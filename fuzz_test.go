package lexy_test

import (
	"bytes"
	"cmp"
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

// translates representations, used for bits<->float
type converter[T, U any] interface {
	to(t T) U
	from(u U) T
	cmp(a, b T) int
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
			return cmp.Compare(bBits, aBits)
		} else {
			// Codec flips the high bit, compare as signed ints
			return cmp.Compare(int64(aBits), int64(bBits))
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

type float32Converter struct{}

func (c float32Converter) to(f float32) uint32   { return math.Float32bits(f) }
func (c float32Converter) from(u uint32) float32 { return math.Float32frombits(u) }
func (c float32Converter) cmp(a, b float32) int {
	return cmpFloats(math.Float32bits, a, b)
}

type float64Converter struct{}

func (c float64Converter) to(f float64) uint64   { return math.Float64bits(f) }
func (c float64Converter) from(u uint64) float64 { return math.Float64frombits(u) }
func (c float64Converter) cmp(a, b float64) int {
	return cmpFloats(math.Float64bits, a, b)
}

func valueTesterForConv[T, U any](codec lexy.Codec[T], conv converter[T, U]) func(*testing.T, U) {
	return func(t *testing.T, repr U) {
		value := conv.from(repr)
		b, err := lexy.Encode(codec, value)
		require.NoError(t, err)
		got, err := lexy.Decode(codec, b)
		require.NoError(t, err)
		assert.IsType(t, value, got)
		assert.Equal(t, conv.to(value), conv.to(got), "values not equal: %#v, %#v", value, got)
	}
}

func FuzzUint8(f *testing.F) {
	addValues(f, seedsUint8...)
	f.Fuzz(valueTesterFor(lexy.Uint8[uint8]()))
}

func FuzzUint16(f *testing.F) {
	addValues(f, seedsUint16...)
	f.Fuzz(valueTesterFor(lexy.Uint16[uint16]()))
}

func FuzzUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(valueTesterFor(lexy.Uint32[uint32]()))
}

func FuzzUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(valueTesterFor(lexy.Uint64[uint64]()))
}

func FuzzInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(valueTesterFor(lexy.Int8[int8]()))
}

func FuzzInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(valueTesterFor(lexy.Int16[int16]()))
}

func FuzzInt32(f *testing.F) {
	addValues(f, seedsInt32...)
	f.Fuzz(valueTesterFor(lexy.Int32[int32]()))
}

func FuzzInt64(f *testing.F) {
	addValues(f, seedsInt64...)
	f.Fuzz(valueTesterFor(lexy.Int64[int64]()))
}

func FuzzFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(valueTesterForConv(lexy.Float32[float32](), float32Converter{}))
}

func FuzzFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(valueTesterForConv(lexy.Float64[float64](), float64Converter{}))
}

func FuzzString(f *testing.F) {
	addValues(f, seedsString...)
	f.Fuzz(valueTesterFor(lexy.String[string]()))
}

func FuzzBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.Bytes[[]byte]()))
}

func FuzzNegUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.Uint32[uint32]())))
}

func FuzzNegInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.Int8[int8]())))
}

func FuzzNegFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(valueTesterForConv(lexy.Negate(lexy.Float64[float64]()), float64Converter{}))
}

func FuzzNegBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.Negate(lexy.Bytes[[]byte]())))
}

func FuzzTerminateUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.Uint64[uint64]())))
}

func FuzzTerminateInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.Int16[int16]())))
}

func FuzzTerminateFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(valueTesterForConv(lexy.Terminate(lexy.Float32[float32]()), float32Converter{}))
}

func FuzzTerminateBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(valueTesterFor(lexy.Terminate(lexy.Bytes[[]byte]())))
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
	f := pairTesterFor(codec, conv.cmp)
	return func(t *testing.T, a, b U) {
		f(t, conv.from(a), conv.from(b))
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
	f.Fuzz(pairTesterFor(lexy.Uint8[uint8](), cmp.Compare[uint8]))
}

func FuzzCmpUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(pairTesterFor(lexy.Uint16[uint16](), cmp.Compare[uint16]))
}

func FuzzCmpUint32(f *testing.F) {
	addUnorderedPairs(f, seedsUint32...)
	f.Fuzz(pairTesterFor(lexy.Uint32[uint32](), cmp.Compare[uint32]))
}

func FuzzCmpUint64(f *testing.F) {
	addUnorderedPairs(f, seedsUint64...)
	f.Fuzz(pairTesterFor(lexy.Uint64[uint64](), cmp.Compare[uint64]))
}

func FuzzCmpInt8(f *testing.F) {
	addUnorderedPairs(f, seedsInt8...)
	f.Fuzz(pairTesterFor(lexy.Int8[int8](), cmp.Compare[int8]))
}

func FuzzCmpInt16(f *testing.F) {
	addUnorderedPairs(f, seedsInt16...)
	f.Fuzz(pairTesterFor(lexy.Int16[int16](), cmp.Compare[int16]))
}

func FuzzCmpInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(pairTesterFor(lexy.Int32[int32](), cmp.Compare[int32]))
}

func FuzzCmpInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(pairTesterFor(lexy.Int64[int64](), cmp.Compare[int64]))
}

func FuzzCmpFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(pairTesterForConv(lexy.Float32[float32](), float32Converter{}))
}

func FuzzCmpFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(pairTesterForConv(lexy.Float64[float64](), float64Converter{}))
}

func FuzzCmpString(f *testing.F) {
	addUnorderedPairs(f, seedsString...)
	f.Fuzz(pairTesterFor(lexy.String[string](), cmp.Compare[string]))
}

func FuzzCmpBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.Bytes[[]byte](), cmpBytes))
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

type negFloat32 struct{}

func (c negFloat32) to(f float32) uint32   { return float32Converter{}.to(negativeFloat32(f)) }
func (c negFloat32) from(u uint32) float32 { return negativeFloat32(float32Converter{}.from(u)) }
func (c negFloat32) cmp(a, b float32) int {
	return float32Converter{}.cmp(b, a)
}

func FuzzCmpNegUint8(f *testing.F) {
	addUnorderedPairs(f, seedsUint8...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.Uint8[uint8]()), negCmp(cmp.Compare[uint8])))
}

func FuzzCmpNegInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.Int32[int32]()), negCmp(cmp.Compare[int32])))
}

func FuzzCmpNegFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(pairTesterForConv(lexy.Negate(lexy.Float32[float32]()), negFloat32{}))
}

func FuzzCmpNegBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.Negate(lexy.Bytes[[]byte]()), negCmp(cmpBytes)))
}

func FuzzCmpTerminateUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.Uint16[uint16]()), cmp.Compare[uint16]))
}

func FuzzCmpTerminateInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.Int64[int64]()), cmp.Compare[int64]))
}

func FuzzCmpTerminateFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(pairTesterForConv(lexy.Terminate(lexy.Float64[float64]()), float64Converter{}))
}

func FuzzCmpTerminateBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(pairTesterFor(lexy.Terminate(lexy.Bytes[[]byte]()), cmpBytes))
}
