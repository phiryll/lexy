package lexy_test

// This file contains things that help in writing Codec tests.
// There are no top-level tests here, but the bulk of
// the Codec-testing code is in testerCodec's methods.

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"testing/iotest"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func encoderFor[T any](codec lexy.Codec[T]) func(value T) []byte {
	return func(value T) []byte {
		return codec.Append(nil, value)
	}
}

func errorForEOF(bytesRead int) error {
	if bytesRead == 0 {
		return io.EOF
	}
	return io.ErrUnexpectedEOF
}

func toCodec[T any](codec lexy.NillableCodec[T]) lexy.Codec[T] {
	return codec
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
//	Write
//	  - Write(testCase.value) == testCase.data
//	  - errors when the io.Writer errors 1 byte before the end of writing
//	  - errors when the io.Writer silently fails to write 1 byte before the end of writing
//	Get
//	  - Get(testCase.data) == testCase.value
//	  - byte count return value == len(testCase.data)
//	  - does not modify the buffer
//	  - Get(testCase.data[:size-1]) either
//	    - panics
//	    - OR if expected value is non-zero, returns the wrong value
//	      - AND byte count return value == size-1
//	Read
//	  - Read(testCase.data) == testCase.value
//	  - consumes len(testCase.data) bytes from the io.Reader
//	  - Read(NewReader(testCase.data[:size-1])) either
//	    - errors with EOF if size-1 == 0, or with ErrUnexpectedEOF otherwise
//	    - OR if expected value is non-zero, returns the wrong value
//	      - AND fully consumes the io.Reader
//
//nolint:thelper
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, true}.test(t, tests)
}

// testVaryingCodec tests Codecs whose encodings may vary for the same input.
// Maps are the only current use case, because of their random iteration order.
//
// This performs all of the same tests as testCodec, except it doesn't use testCase.data.
// Instead, it tests each of the outputs of Append/Put/Write as inputs to Get/Read, all combinations.
// It also tests that different ways of invoking Append/Put/Write always output the same number of bytes.
//
//nolint:thelper
func testVaryingCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, false}.test(t, tests)
}

// A testing wrapper for Codec that deals only in []buf at the API level.
type testerCodec[T any] struct {
	codec        lexy.Codec[T]
	isConsistent bool
}

// The output of one Append/Put/Write method.
type output struct {
	name string
	buf  []byte
}

//nolint:thelper
func (c testerCodec[T]) test(t *testing.T, tests []testCase[T]) {
	// Test Get/Read behavior with empty or error inputs.
	c.getEmpty(t)
	c.readEmpty(t)
	c.readError(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("Test case: %v", tt)

			// Test Put/Write to destinations that are too short by 1 byte.
			c.putShortBuf(t, tt)
			c.writeShortBuf(t, tt)
			c.writeSilentError(t, tt)

			// Test Append/Put/Write.
			var outputs []output
			outputs = append(outputs, c.appendNil(t, tt))
			outputs = append(outputs, c.appendExisting(t, tt))
			outputs = append(outputs, c.put(t, tt))
			outputs = append(outputs, c.putLongBuf(t, tt))
			outputs = append(outputs, c.write(t, tt))
			t.Logf("Outputs: %v", outputs)

			// Test Get/Read
			if c.isConsistent {
				c.get(t, tt, tt.data)
				c.getShortBuf(t, tt, tt.data)
				c.read(t, tt, tt.data)
				c.readShortBuf(t, tt, tt.data)
			} else {
				for _, out := range outputs {
					out := out
					t.Run("round trip: "+out.name+" to", func(t *testing.T) {
						t.Parallel()
						c.get(t, tt, out.buf)
						c.getShortBuf(t, tt, out.buf)
						c.read(t, tt, out.buf)
						c.readShortBuf(t, tt, out.buf)
					})
				}
			}
		})
	}
}

//nolint:thelper
func (c testerCodec[T]) getEmpty(t *testing.T) {
	t.Run("get empty", func(t *testing.T) {
		var zero T
		got, gotSize := c.codec.Get([]byte{})
		if gotSize != -1 {
			assert.Equal(t, 0, gotSize)
		}
		// All Codecs in lexy can only return the zero value from zero bytes.
		assert.IsType(t, zero, got)
		assert.Equal(t, zero, got)
	})
}

//nolint:thelper
func (c testerCodec[T]) readEmpty(t *testing.T) {
	t.Run("read empty", func(t *testing.T) {
		var zero T
		r := bytes.NewReader([]byte{})
		got, err := c.codec.Read(r)
		if err != nil {
			assert.ErrorIs(t, err, io.EOF)
		}
		// All Codecs in lexy can only return the zero value from empty input.
		assert.IsType(t, zero, got)
		assert.Equal(t, zero, got)
	})
}

//nolint:thelper
func (c testerCodec[T]) readError(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		var zero T
		r := iotest.ErrReader(errRead)
		got, err := c.codec.Read(r)
		// emptyCodec never errors, and there's no good way to not include it.
		if err != nil {
			assert.ErrorIs(t, err, errRead)
		}
		assert.IsType(t, zero, got)
		assert.Equal(t, zero, got)
	})
}

//nolint:thelper
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

//nolint:thelper
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
			assert.Equal(t, size, len(buf))
		}
	})
	return output{"append existing", buf}
}

//nolint:thelper
func (c testerCodec[T]) put(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		buf = make([]byte, size)
		putSize := c.codec.Put(buf, tt.value)
		assert.Equal(t, size, putSize)
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put", buf}
}

//nolint:thelper
func (c testerCodec[T]) putLongBuf(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put long buf", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		buf = make([]byte, size+1000)
		for i := range buf {
			buf[i] = 37
		}
		putSize := c.codec.Put(buf, tt.value)
		assert.Equal(t, size, putSize)
		for i := range buf[size:] {
			k := size + i
			assert.Equal(t, byte(37), buf[k], "buf[%d] = %d written to buffer", k, buf[k])
		}
		buf = buf[:putSize]
		if c.isConsistent {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put long buf", buf}
}

//nolint:thelper
func (c testerCodec[T]) putShortBuf(t *testing.T, tt testCase[T]) {
	t.Run("put short buf", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := make([]byte, size+100)
		assert.Panics(t, func() {
			c.codec.Put(buf[:size-1], tt.value)
		})
	})
}

//nolint:thelper
func (c testerCodec[T]) write(t *testing.T, tt testCase[T]) output {
	buf := bytes.NewBuffer([]byte{})
	t.Run("write", func(t *testing.T) {
		err := c.codec.Write(buf, tt.value)
		require.NoError(t, err)
		if c.isConsistent {
			assert.Equal(t, tt.data, buf.Bytes())
		} else {
			size := len(c.codec.Append(nil, tt.value))
			assert.Equal(t, size, buf.Len())
		}
	})
	return output{"write", buf.Bytes()}
}

//nolint:thelper
func (c testerCodec[T]) writeShortBuf(t *testing.T, tt testCase[T]) {
	t.Run("write short buf", func(t *testing.T) {
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := bytes.NewBuffer([]byte{})
		w := boundedWriter{buf, size - 1}
		err := c.codec.Write(&w, tt.value)
		require.Error(t, err)
	})
}

//nolint:thelper
func (c testerCodec[T]) writeSilentError(t *testing.T, tt testCase[T]) {
	t.Run("write silent error", func(t *testing.T) {
		t.Skip("TODO: Write does not yet fail properly on silent truncation.")
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := bytes.NewBuffer([]byte{})
		err := c.codec.Write(iotest.TruncateWriter(buf, int64(size-1)), tt.value)
		require.Error(t, err)
	})
}

//nolint:thelper
func (c testerCodec[T]) get(t *testing.T, tt testCase[T], buf []byte) {
	workingBuf := append([]byte{}, buf...)
	t.Run("get", func(t *testing.T) {
		got, gotSize := c.codec.Get(workingBuf)
		var expected T
		if gotSize == -1 {
			assert.Equal(t, 0, len(buf))
		} else {
			expected = tt.value
			assert.Equal(t, len(buf), gotSize)
		}
		assert.IsType(t, expected, got)
		assert.Equal(t, expected, got)
		assert.Equal(t, buf, workingBuf)
	})
}

//nolint:thelper
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
		var gotSize int
		panicked := func() (panicked bool) {
			panicked = false
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			got, gotSize = c.codec.Get(workingBuf[:size-1])
			return
		}()
		if !panicked {
			assert.Equal(t, size-1, gotSize, "got too much data")
			assert.IsType(t, tt.value, got)
			if !reflect.ValueOf(tt.value).IsZero() { // both might be randomly zero
				assert.NotEqual(t, tt.value, got, "got value without full data")
			}
		}
		assert.Equal(t, buf, workingBuf)
	})
}

//nolint:thelper
func (c testerCodec[T]) read(t *testing.T, tt testCase[T], buf []byte) {
	workingBuf := append([]byte{}, buf...)
	t.Run("read", func(t *testing.T) {
		r := bytes.NewReader(workingBuf)
		got, err := c.codec.Read(r)
		require.NoError(t, err)
		assert.Equal(t, 0, r.Len())
		assert.IsType(t, tt.value, got)
		assert.Equal(t, tt.value, got)
		assert.Equal(t, buf, workingBuf)
	})
}

//nolint:thelper
func (c testerCodec[T]) readShortBuf(t *testing.T, tt testCase[T], buf []byte) {
	workingBuf := append([]byte{}, buf...)
	t.Run("read short buf", func(t *testing.T) {
		size := len(buf)
		if size == 0 {
			return
		}
		// Should either error, or read one fewer byte and get the wrong value back.
		r := bytes.NewReader(workingBuf[:size-1])
		got, err := c.codec.Read(r)
		if err != nil {
			assert.ErrorIs(t, err, errorForEOF(size-1))
		}
		assert.IsType(t, tt.value, got)
		if !reflect.ValueOf(tt.value).IsZero() { // both might be randomly zero
			assert.NotEqual(t, tt.value, got, "got value without full data")
		}
		assert.Equal(t, 0, r.Len())
		assert.Equal(t, buf, workingBuf)
	})
}
