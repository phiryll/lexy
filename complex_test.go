package lexy_test

import (
	"fmt"
	"testing"

	"github.com/phiryll/lexy"
)

func comp64(r, i float32) complex64   { return complex(r, i) }
func comp128(r, i float64) complex128 { return complex(r, i) }

func pairTestCases[T, P any](tests []testCase[T], pair func(a, b T) P) []testCase[P] {
	result := []testCase[P]{}
	for _, a := range tests {
		for _, b := range tests {
			result = append(result, testCase[P]{
				fmt.Sprintf("(%s, %s)", a.name, b.name),
				pair(a.value, b.value),
				nil,
			})
		}
	}
	return result
}

func TestComplex64(t *testing.T) {
	t.Parallel()
	codec := lexy.Complex64()
	testCodec(t, codec, fillTestData(codec, pairTestCases(float32NumberTestCases, comp64)))
}

func TestComplex64Ordering(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.Complex64(), pairTestCases(float32TestCases, comp64))
}

func TestComplex128(t *testing.T) {
	t.Parallel()
	codec := lexy.Complex128()
	testCodec(t, codec, fillTestData(codec, pairTestCases(float64NumberTestCases, comp128)))
}

func TestComplex128Ordering(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.Complex128(), pairTestCases(float64TestCases, comp128))
}
