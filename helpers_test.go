package lexy_test

// This file contains things that help in writing Codec tests.
// There are no top-level tests here, but the bulk of
// the Codec-testing code is in testerCodec's methods.

import (
	"reflect"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

// Just to make the test cases terser.
const (
	term      byte = lexy.TestingTerminator
	esc       byte = lexy.TestingEscape
	pNilFirst byte = lexy.TestingPrefixNilFirst
	pNonNil   byte = lexy.TestingPrefixNonNil
	pNilLast  byte = lexy.TestingPrefixNilLast
)

func ptr[T any](value T) *T {
	return &value
}

func concat(slices ...[]byte) []byte {
	var result []byte
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

type testCase[T any] struct {
	name  string
	value T
	data  []byte
}

// Returns new test cases with each testCase.data set to codec.Append(nil, testCase.value).
// This is useful when the encoded value is difficult to calculate by hand.
func fillTestData[T any](codec lexy.Codec[T], tests []testCase[T]) []testCase[T] {
	newTests := make([]testCase[T], len(tests))
	for i, tt := range tests {
		test := tt
		test.data = codec.Append(nil, tt.value)
		newTests[i] = test
	}
	return newTests
}

// testCodec tests:
//
//	Append
//	  - Append(testCase.value) == testCase.data
//	  - does not modify the buffer's existing data
//	Put
//	  - Put(testCase.value) == testCase.data
//	  - byte count return value == len(testCase.data)
//	  - does not modify the buffer's existing data beyond the value written
//	  - panics when the buffer is 1 byte too short
//	Get
//	  - Get(testCase.data) == testCase.value
//	  - byte count return value == len(testCase.data)
//	  - does not modify the buffer
//	  - Get(testCase.data[:size-1]) either
//	    - panics
//	    - OR if expected value is non-zero, returns the wrong value
//	      - AND byte count return value == size-1
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, true}.test(t, tests)
}

// testVaryingCodec tests Codecs whose encodings may vary for the same input.
// Maps are the only current use case, because of their random iteration order.
//
// This performs all of the same tests as testCodec, except it doesn't use testCase.data.
// Instead, it tests each of the outputs of Append/Put as inputs to Get, all combinations.
// It also tests that different ways of invoking Append/Put always output the same number of bytes.
func testVaryingCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, false}.test(t, tests)
}

// A testing wrapper for Codec that deals only in []buf at the API level.
type testerCodec[T any] struct {
	codec        lexy.Codec[T]
	isConsistent bool
}

// The output of one Append/Put method.
type output struct {
	name string
	buf  []byte
}

func (c testerCodec[T]) test(t *testing.T, tests []testCase[T]) {
	c.getEmpty(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("Test case: %v", tt)

			c.putShortBuf(t, tt)

			// Test Append/Put
			outputs := append([]output{},
				c.appendNil(t, tt),
				c.appendExisting(t, tt),
				c.put(t, tt),
				c.putLongBuf(t, tt))
			t.Logf("Outputs: %v", outputs)

			// Test Get
			if c.isConsistent {
				c.get(t, tt, tt.data)
				c.getShortBuf(t, tt, tt.data)
			} else {
				for _, out := range outputs {
					out := out
					t.Run("round trip: "+out.name+" to", func(t *testing.T) {
						t.Parallel()
						c.get(t, tt, out.buf)
						c.getShortBuf(t, tt, out.buf)
					})
				}
			}
		})
	}
}

//nolint:revive
func (c testerCodec[T]) getEmpty(t *testing.T) {
	t.Run("get empty", func(t *testing.T) {
		var zero T
		if len(c.codec.Append([]byte{}, zero)) == 0 {
			return
		}
		assert.Panics(t, func() {
			c.codec.Get([]byte{})
		})
	})
}

func (c testerCodec[T]) appendNil(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("append nil", func(t *testing.T) {
		buf = c.codec.Append(nil, tt.value)
		if buf == nil {
			buf = []byte{}
		}
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"append nil", buf}
}

func (c testerCodec[T]) appendExisting(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("append existing", func(t *testing.T) {
		header := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buf = append(buf, header...)
		buf = c.codec.Append(buf, tt.value)
		assert.Equal(t, header, buf[:len(header)])
		buf = buf[len(header):]
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		} else {
			size := len(c.codec.Append(nil, tt.value))
			assert.Len(t, buf, size)
		}
	})
	return output{"append existing", buf}
}

func (c testerCodec[T]) put(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		buf = make([]byte, size)
		bufAfter := c.codec.Put(buf, tt.value)
		assert.Empty(t, bufAfter)
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put", buf}
}

func (c testerCodec[T]) putLongBuf(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put long buf", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		buf = make([]byte, size+10)
		for i := range buf {
			buf[i] = 37
		}
		bufAfter := c.codec.Put(buf, tt.value)
		putSize := len(buf) - len(bufAfter)
		assert.Equal(t, size, putSize)
		for i := range buf[size:] {
			k := size + i
			assert.Equal(t, byte(37), buf[k], "buf[%d] = %d written to buffer", k, buf[k])
		}
		buf = buf[:size]
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put long buf", buf}
}

func (c testerCodec[T]) putShortBuf(t *testing.T, tt testCase[T]) {
	t.Run("put short buf", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := make([]byte, size+2000)
		assert.Panics(t, func() {
			c.codec.Put(buf[:size-1], tt.value)
		})
	})
}

func (c testerCodec[T]) get(t *testing.T, tt testCase[T], buf []byte) {
	workingBuf := append([]byte{}, buf...)
	t.Run("get", func(t *testing.T) {
		got, gotBuf := c.codec.Get(workingBuf)
		// Only empty because that's how these tests are set up.
		assert.Empty(t, len(gotBuf))
		assert.IsType(t, tt.value, got)
		assert.Equal(t, tt.value, got)
		assert.Equal(t, buf, workingBuf)
	})
}

//nolint:revive
func (c testerCodec[T]) getShortBuf(t *testing.T, tt testCase[T], buf []byte) {
	workingBuf := append([]byte{}, buf...)
	t.Run("get short buf", func(t *testing.T) {
		size := len(buf)
		if size <= 1 {
			// shortening 1 to 0 results in a special case
			return
		}
		// Should either panic, or read one fewer byte and get the wrong value back.
		var got T
		var gotBuf []byte
		//nolint:nakedret,nonamedreturns
		panicked := func() (panicked bool) {
			panicked = false
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			got, gotBuf = c.codec.Get(workingBuf[:size-1])
			return
		}()
		if !panicked {
			assert.Empty(t, gotBuf, "got wrong amount of data")
			assert.IsType(t, tt.value, got)
			if !reflect.ValueOf(tt.value).IsZero() { // both might be randomly zero
				assert.NotEqual(t, tt.value, got, "got value without full data")
			}
		}
		assert.Equal(t, buf, workingBuf)
	})
}

func testOrdering[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	tests = fillTestData(codec, tests)
	for i := range tests {
		if i == 0 {
			continue
		}
		a := tests[i-1]
		b := tests[i]
		t.Run(a.name+" < "+b.name, func(t *testing.T) {
			t.Parallel()
			assert.Less(t, a.data, b.data)
		})
	}
}
