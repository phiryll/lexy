package lexy_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reusing things from float_test.

// float32s in increasing order.
var float32s = []struct {
	name  string
	value float32
}{
	{"-max NaN", negMaxNaN32},
	{"-min NaN", negMinNaN32},
	{"-Inf", negInf32},
	{"-max normal", negMaxNormal32},
	{"-min normal", negMinNormal32},
	{"-max subnormal", negMaxSubnormal32},
	{"-min subnormal", negMinSubnormal32},
	{"-0", negZero32},
	{"+0", posZero32},
	{"+min subnormal", posMinSubnormal32},
	{"+max subnormal", posMaxSubnormal32},
	{"+min normal", posMinNormal32},
	{"+max normal", posMaxNormal32},
	{"+Inf", posInf32},
	{"+min NaN", posMinNaN32},
	{"+max NaN", posMaxNaN32},
}

func TestComplex64(t *testing.T) {
	codec := lexy.Complex64()
	// Ensure we don't get complex128 without having to cast the arguments.
	comp := func(r, i float32) complex64 { return complex(r, i) }

	// all pairs of float32s in increasing "order"
	testCases := []testCase[complex64]{}
	for _, r := range float32s {
		for _, i := range float32s {
			testCases = append(testCases, testCase[complex64]{
				fmt.Sprintf("(%s, %s)", r.name, i.name),
				comp(r.value, i.value),
				nil,
			})
		}
	}

	// test round trip
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			err := codec.Write(buf, tt.value)
			require.NoError(t, err)
			got, err := codec.Read(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err)
			// works for NaN as well
			assert.Equal(t, math.Float32bits(real(tt.value)), math.Float32bits(real(got)))
		})
	}

	// test ordering
	var prev []byte
	for i, tt := range testCases {
		name := tt.name
		if i > 0 {
			name = fmt.Sprintf("%s < %s", testCases[i-1].name, name)
		}
		t.Run(name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			err := codec.Write(buf, tt.value)
			require.NoError(t, err)
			current := buf.Bytes()
			if i > 0 {
				assert.Less(t, prev, current)
			}
			prev = current
		})
	}
}

// float64s in increasing order.
var float64s = []struct {
	name  string
	value float64
}{
	{"-max NaN", negMaxNaN64},
	{"-min NaN", negMinNaN64},
	{"-Inf", negInf64},
	{"-max normal", negMaxNormal64},
	{"-min normal", negMinNormal64},
	{"-max subnormal", negMaxSubnormal64},
	{"-min subnormal", negMinSubnormal64},
	{"-0", negZero64},
	{"+0", posZero64},
	{"+min subnormal", posMinSubnormal64},
	{"+max subnormal", posMaxSubnormal64},
	{"+min normal", posMinNormal64},
	{"+max normal", posMaxNormal64},
	{"+Inf", posInf64},
	{"+min NaN", posMinNaN64},
	{"+max NaN", posMaxNaN64},
}

func TestComplex128(t *testing.T) {
	codec := lexy.Complex128()
	// Ensure we don't get complex128 without having to cast the arguments.
	comp := func(r, i float64) complex128 { return complex(r, i) }

	// all pairs of float64s in increasing "order"
	testCases := []testCase[complex128]{}
	for _, r := range float64s {
		for _, i := range float64s {
			testCases = append(testCases, testCase[complex128]{
				fmt.Sprintf("(%s, %s)", r.name, i.name),
				comp(r.value, i.value),
				nil,
			})
		}
	}

	// test round trip
	for _, tt := range testCases {
		buf := bytes.NewBuffer([]byte{})
		err := codec.Write(buf, tt.value)
		require.NoError(t, err)
		got, err := codec.Read(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		// works for NaN as well
		assert.Equal(t, math.Float64bits(real(tt.value)), math.Float64bits(real(got)))
	}

	// test ordering
	var prev []byte
	for i, tt := range testCases {
		name := tt.name
		if i > 0 {
			name = fmt.Sprintf("%s < %s", testCases[i-1].name, name)
		}
		t.Run(name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			err := codec.Write(buf, tt.value)
			require.NoError(t, err)
			current := buf.Bytes()
			if i > 0 {
				assert.Less(t, prev, current)
			}
			prev = current
		})
	}
}
