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
)

// These fuzzers test the encode-decode round trip.

func addValues[T any](f *testing.F, values ...T) {
	for _, x := range values {
		f.Add(x)
	}
}

func valueTesterFor[T any](codec lexy.Codec[T]) func(t *testing.T, value T) {
	return func(t *testing.T, value T) {
		b, err := lexy.Encode(codec, value)
		require.NoError(t, err)
		got, err := lexy.Decode(codec, b)
		require.NoError(t, err)
		assert.IsType(t, value, got)
		assert.Equal(t, value, got)
	}
}

func FuzzUint8(f *testing.F) {
	addValues(f, seedsUint8...)
	f.Fuzz(valueTesterFor(lexy.Uint[uint8]()))
}

func FuzzUint16(f *testing.F) {
	addValues(f, seedsUint16...)
	f.Fuzz(valueTesterFor(lexy.Uint[uint16]()))
}

func FuzzUint32(f *testing.F) {
	addValues(f, seedsUint32...)
	f.Fuzz(valueTesterFor(lexy.Uint[uint32]()))
}

func FuzzUint64(f *testing.F) {
	addValues(f, seedsUint64...)
	f.Fuzz(valueTesterFor(lexy.Uint[uint64]()))
}

func FuzzInt8(f *testing.F) {
	addValues(f, seedsInt8...)
	f.Fuzz(valueTesterFor(lexy.Int[int8]()))
}

func FuzzInt16(f *testing.F) {
	addValues(f, seedsInt16...)
	f.Fuzz(valueTesterFor(lexy.Int[int16]()))
}

func FuzzInt32(f *testing.F) {
	addValues(f, seedsInt32...)
	f.Fuzz(valueTesterFor(lexy.Int[int32]()))
}

func FuzzInt64(f *testing.F) {
	addValues(f, seedsInt64...)
	f.Fuzz(valueTesterFor(lexy.Int[int64]()))
}

// These fuzzers test that the encoding order is consistent with the value order.

func addPairs[T any](f *testing.F, values ...T) {
	for i, x := range values {
		for _, y := range values[i+1:] {
			f.Add(x, y)
		}
	}
}

func pairTesterFor[T any](codec lexy.Codec[T], cmp func(a, b T) int) func(t *testing.T, a, b T) {
	return func(t *testing.T, a, b T) {
		aEncoded, err := lexy.Encode(codec, a)
		require.NoError(t, err)
		bEncoded, err := lexy.Encode(codec, b)
		require.NoError(t, err)
		assert.Equal(t, cmp(a, b), bytes.Compare(aEncoded, bEncoded))
	}
}

func FuzzCmpUint8(f *testing.F) {
	addPairs(f, seedsUint8...)
	f.Fuzz(pairTesterFor(lexy.Uint[uint8](), cmp.Compare[uint8]))
}

func FuzzCmpUint16(f *testing.F) {
	addPairs(f, seedsUint16...)
	f.Fuzz(pairTesterFor(lexy.Uint[uint16](), cmp.Compare[uint16]))
}

func FuzzCmpUint32(f *testing.F) {
	addPairs(f, seedsUint32...)
	f.Fuzz(pairTesterFor(lexy.Uint[uint32](), cmp.Compare[uint32]))
}

func FuzzCmpUint64(f *testing.F) {
	addPairs(f, seedsUint64...)
	f.Fuzz(pairTesterFor(lexy.Uint[uint64](), cmp.Compare[uint64]))
}

func FuzzCmpInt8(f *testing.F) {
	addPairs(f, seedsInt8...)
	f.Fuzz(pairTesterFor(lexy.Int[int8](), cmp.Compare[int8]))
}

func FuzzCmpInt16(f *testing.F) {
	addPairs(f, seedsInt16...)
	f.Fuzz(pairTesterFor(lexy.Int[int16](), cmp.Compare[int16]))
}

func FuzzCmpInt32(f *testing.F) {
	addPairs(f, seedsInt32...)
	f.Fuzz(pairTesterFor(lexy.Int[int32](), cmp.Compare[int32]))
}

func FuzzCmpInt64(f *testing.F) {
	addPairs(f, seedsInt64...)
	f.Fuzz(pairTesterFor(lexy.Int[int64](), cmp.Compare[int64]))
}
