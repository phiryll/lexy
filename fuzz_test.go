package lexy_test

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
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
	// because Go's float fuzzer only generates one pattern for NaN.
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

// Comparison functions.

// Negates a comparison function.
func negCmp[T any](cmp func(T, T) int) func(T, T) int {
	return func(a, b T) int {
		return cmp(b, a)
	}
}

// Helper function somewhat duplicating cmp.Compare (Go 1.21).
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

// This is necessary because bytes.Compare(nil, {}) == 0.
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

// Implements ordering semantics of the float Codecs, mostly without encoding them.
func cmpFloats[T float32 | float64, U uint32 | uint64](aBits, bBits U, a, b T) int {
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

func cmpUintFloat32(a, b uint32) int {
	return cmpFloats(a, b, math.Float32frombits(a), math.Float32frombits(b))
}

func cmpUintFloat64(a, b uint64) int {
	return cmpFloats(a, b, math.Float64frombits(a), math.Float64frombits(b))
}

// Codecs that translate representations, used for uint bits<->float.

func toUint32(codec lexy.Codec[float32]) lexy.Codec[uint32] {
	return toUint32Codec{codec}
}

func toUint64(codec lexy.Codec[float64]) lexy.Codec[uint64] {
	return toUint64Codec{codec}
}

type toUint32Codec struct {
	codec lexy.Codec[float32]
}

func (c toUint32Codec) Append(buf []byte, value uint32) []byte {
	return c.codec.Append(buf, math.Float32frombits(value))
}

func (c toUint32Codec) Put(buf []byte, value uint32) int {
	return c.codec.Put(buf, math.Float32frombits(value))
}

func (c toUint32Codec) Get(buf []byte) (uint32, int) {
	value, n := c.codec.Get(buf)
	return math.Float32bits(value), n
}

func (c toUint32Codec) Write(w io.Writer, value uint32) error {
	return c.codec.Write(w, math.Float32frombits(value))
}

func (c toUint32Codec) Read(r io.Reader) (uint32, error) {
	value, err := c.codec.Read(r)
	return math.Float32bits(value), err
}

func (toUint32Codec) RequiresTerminator() bool {
	return false
}

type toUint64Codec struct {
	codec lexy.Codec[float64]
}

func (c toUint64Codec) Append(buf []byte, value uint64) []byte {
	return c.codec.Append(buf, math.Float64frombits(value))
}

func (c toUint64Codec) Put(buf []byte, value uint64) int {
	return c.codec.Put(buf, math.Float64frombits(value))
}

func (c toUint64Codec) Get(buf []byte) (uint64, int) {
	value, n := c.codec.Get(buf)
	return math.Float64bits(value), n
}

func (c toUint64Codec) Write(w io.Writer, value uint64) error {
	return c.codec.Write(w, math.Float64frombits(value))
}

func (c toUint64Codec) Read(r io.Reader) (uint64, error) {
	value, err := c.codec.Read(r)
	return math.Float64bits(value), err
}

func (toUint64Codec) RequiresTerminator() bool {
	return false
}

// Functions to add seed values to the fuzzer.

func addValues[T any](f *testing.F, values ...T) {
	f.Helper()
	for _, x := range values {
		f.Add(x)
	}
}

// Used for testing cmp(v1, v2).
func addUnorderedPairs[T any](f *testing.F, values ...T) {
	f.Helper()
	for i, x := range values {
		for _, y := range values[i+1:] {
			f.Add(x, y)
		}
	}
}

// Functions to create fuzz targets.

func fuzzTargetForValue[T any](codec lexy.Codec[T]) func(*testing.T, T) {
	//nolint:thelper
	return func(t *testing.T, value T) {
		testCodec(t, codec, []testCase[T]{
			{"fuzz", value, codec.Append([]byte{}, value)},
		})
	}
}

func fuzzTargetForPair[T any](codec lexy.Codec[T], cmp func(T, T) int) func(*testing.T, T, T) {
	//nolint:thelper
	return func(t *testing.T, a, b T) {
		aEncoded := codec.Append(nil, a)
		bEncoded := codec.Append(nil, b)
		assert.Equal(t, cmp(a, b), bytes.Compare(aEncoded, bEncoded),
			"values not comparing correctly: %#v(%x), %#v(%x)", a, aEncoded, b, bEncoded)
	}
}

func FuzzUint8(f *testing.F) {
	addValues(f, seedsUint8...)
	f.Fuzz(fuzzTargetForValue(lexy.Uint8()))
}

func FuzzUint16(f *testing.F) {
	addValues(f, seedsUint16...)
	f.Fuzz(fuzzTargetForValue(lexy.Uint16()))
}

func FuzzUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(fuzzTargetForValue(lexy.Uint32()))
}

func FuzzUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(fuzzTargetForValue(lexy.Uint64()))
}

func FuzzInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(fuzzTargetForValue(lexy.Int8()))
}

func FuzzInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(fuzzTargetForValue(lexy.Int16()))
}

func FuzzInt32(f *testing.F) {
	addValues(f, seedsInt32...)
	f.Fuzz(fuzzTargetForValue(lexy.Int32()))
}

func FuzzInt64(f *testing.F) {
	addValues(f, seedsInt64...)
	f.Fuzz(fuzzTargetForValue(lexy.Int64()))
}

func FuzzFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(fuzzTargetForValue(toUint32(lexy.Float32())))
}

func FuzzFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(fuzzTargetForValue(toUint64(lexy.Float64())))
}

func FuzzString(f *testing.F) {
	addValues(f, seedsString...)
	f.Fuzz(fuzzTargetForValue(lexy.String()))
}

func FuzzBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(fuzzTargetForValue(toCodec(lexy.Bytes())))
}

func FuzzNegUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(fuzzTargetForValue(lexy.Negate(lexy.Uint32())))
}

func FuzzNegInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(fuzzTargetForValue(lexy.Negate(lexy.Int8())))
}

func FuzzNegFloat64(f *testing.F) {
	addValues(f, seedsFloat64...)
	f.Fuzz(fuzzTargetForValue(toUint64(lexy.Negate(lexy.Float64()))))
}

func FuzzNegBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(fuzzTargetForValue(lexy.Negate(toCodec(lexy.Bytes()))))
}

func FuzzTerminateUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(fuzzTargetForValue(lexy.Terminate(lexy.Uint64())))
}

func FuzzTerminateInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(fuzzTargetForValue(lexy.Terminate(lexy.Int16())))
}

func FuzzTerminateFloat32(f *testing.F) {
	addValues(f, seedsFloat32...)
	f.Fuzz(fuzzTargetForValue(toUint32(lexy.Terminate(lexy.Float32()))))
}

func FuzzTerminateBytes(f *testing.F) {
	addValues(f, seedsBytes...)
	f.Fuzz(fuzzTargetForValue(lexy.Terminate(toCodec(lexy.Bytes()))))
}

func FuzzCmpUint8(f *testing.F) {
	addUnorderedPairs(f, seedsUint8...)
	f.Fuzz(fuzzTargetForPair(lexy.Uint8(), compare[uint8]))
}

func FuzzCmpUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(fuzzTargetForPair(lexy.Uint16(), compare[uint16]))
}

func FuzzCmpUint32(f *testing.F) {
	addUnorderedPairs(f, seedsUint32...)
	f.Fuzz(fuzzTargetForPair(lexy.Uint32(), compare[uint32]))
}

func FuzzCmpUint64(f *testing.F) {
	addUnorderedPairs(f, seedsUint64...)
	f.Fuzz(fuzzTargetForPair(lexy.Uint64(), compare[uint64]))
}

func FuzzCmpInt8(f *testing.F) {
	addUnorderedPairs(f, seedsInt8...)
	f.Fuzz(fuzzTargetForPair(lexy.Int8(), compare[int8]))
}

func FuzzCmpInt16(f *testing.F) {
	addUnorderedPairs(f, seedsInt16...)
	f.Fuzz(fuzzTargetForPair(lexy.Int16(), compare[int16]))
}

func FuzzCmpInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(fuzzTargetForPair(lexy.Int32(), compare[int32]))
}

func FuzzCmpInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(fuzzTargetForPair(lexy.Int64(), compare[int64]))
}

func FuzzCmpFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(fuzzTargetForPair(toUint32(lexy.Float32()), cmpUintFloat32))
}

func FuzzCmpFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(fuzzTargetForPair(toUint64(lexy.Float64()), cmpUintFloat64))
}

func FuzzCmpString(f *testing.F) {
	addUnorderedPairs(f, seedsString...)
	f.Fuzz(fuzzTargetForPair(lexy.String(), compare[string]))
}

func FuzzCmpBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(fuzzTargetForPair(toCodec(lexy.Bytes()), cmpBytes))
}

func FuzzCmpNegUint8(f *testing.F) {
	addUnorderedPairs(f, seedsUint8...)
	f.Fuzz(fuzzTargetForPair(lexy.Negate(lexy.Uint8()), negCmp(compare[uint8])))
}

func FuzzCmpNegInt32(f *testing.F) {
	addUnorderedPairs(f, seedsInt32...)
	f.Fuzz(fuzzTargetForPair(lexy.Negate(lexy.Int32()), negCmp(compare[int32])))
}

func FuzzCmpNegFloat32(f *testing.F) {
	addUnorderedPairs(f, seedsFloat32...)
	f.Fuzz(fuzzTargetForPair(toUint32(lexy.Negate(lexy.Float32())), negCmp(cmpUintFloat32)))
}

func FuzzCmpNegBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(fuzzTargetForPair(lexy.Negate(toCodec(lexy.Bytes())), negCmp(cmpBytes)))
}

func FuzzCmpTerminateUint16(f *testing.F) {
	addUnorderedPairs(f, seedsUint16...)
	f.Fuzz(fuzzTargetForPair(lexy.Terminate(lexy.Uint16()), compare[uint16]))
}

func FuzzCmpTerminateInt64(f *testing.F) {
	addUnorderedPairs(f, seedsInt64...)
	f.Fuzz(fuzzTargetForPair(lexy.Terminate(lexy.Int64()), compare[int64]))
}

func FuzzCmpTerminateFloat64(f *testing.F) {
	addUnorderedPairs(f, seedsFloat64...)
	f.Fuzz(fuzzTargetForPair(toUint64(lexy.Terminate(lexy.Float64())), cmpUintFloat64))
}

func FuzzCmpTerminateBytes(f *testing.F) {
	addUnorderedPairs(f, seedsBytes...)
	f.Fuzz(fuzzTargetForPair(lexy.Terminate(toCodec(lexy.Bytes())), cmpBytes))
}
