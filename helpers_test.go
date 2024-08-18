package lexy_test

// This file contains things that help in writing Codec tests,
// it doesn't have any tests itself.

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](value T) *T {
	return &value
}

func encoderFor[T any](codec lexy.Codec[T]) func(value T) []byte {
	return func(value T) []byte {
		return codec.Append(nil, value)
	}
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

// Just to make the test cases terser.
const (
	term      byte = lexy.TestingTerminator
	esc       byte = lexy.TestingEscape
	pNilFirst byte = lexy.TestingPrefixNilFirst
	pNonNil   byte = lexy.TestingPrefixNonNil
	pNilLast  byte = lexy.TestingPrefixNilLast
)

// Tests:
// - Append/Put/Write(testCase.value) == testCase.data
// - Get/Read(testCase.data) == testCase.value
// - len(Get/Put byte count return value) == len(testCase.data)
// - Get/Put panic when given a buffer that is 1 byte too short, or return incorrect values.
// - Read errs when given a stream that is 1 byte too short, or returns incorrect values.
//
//nolint:thelper
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("test codec", func(t *testing.T) {
		testCodecAppend(t, codec, tests)
		testCodecPut(t, codec, tests)
		testCodecGet(t, codec, tests)
		testCodecWrite(t, codec, tests)
		testCodecRead(t, codec, tests)
	})
}

// Calculates and sets testCase.data for each of the tests using codec.Append, and then calls testCodec.
// This is useful when the encoded value is difficult to calculate by hand.
//
//nolint:thelper
func testCodecMakeData[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	newTests := make([]testCase[T], len(tests))
	for i, tt := range tests {
		test := tt
		test.data = codec.Append(nil, tt.value)
		newTests[i] = test
	}
	testCodec(t, codec, newTests)
}

//nolint:thelper
func testCodecAppend[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("append nil", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				buf := codec.Append(nil, tt.value)
				if buf == nil {
					buf = []byte{}
				}
				assert.Equal(t, tt.data, buf)
			})
		}
	})
	t.Run("append existing", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				buf := make([]byte, 10)
				buf = codec.Append(buf, tt.value)
				assert.Equal(t, tt.data, buf[10:])
			})
		}
	})
}

//nolint:thelper
func testCodecPut[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("put", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(codec.Append(nil, tt.value))
				buf := make([]byte, size)
				putSize := codec.Put(buf, tt.value)
				assert.Equal(t, size, putSize)
				assert.Equal(t, tt.data, buf)
			})
		}
	})
	t.Run("put short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(codec.Append(nil, tt.value))
				if size == 0 {
					return
				}
				// allocate more than enough space,
				// but limit the size of the sub-slice passed in.
				buf := make([]byte, size+10000)
				assert.Panics(t, func() {
					codec.Put(buf[:size-1], tt.value)
				})
			})
		}
	})
}

//nolint:thelper
func testCodecGet[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("get", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				got, gotSize := codec.Get(tt.data)
				assert.Equal(t, len(tt.data), gotSize)
				assert.IsType(t, tt.value, got)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("get short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(tt.data)
				if size == 0 {
					return
				}
				// Should either panic, or read one fewer byte and get the wrong value back.
				//nolint:errcheck
				defer func() { recover() }()
				got, gotSize := codec.Get(tt.data[:size-1])
				assert.Equal(t, size-1, gotSize, "read too much data")
				assert.NotEqual(t, tt.value, got, "read value without full data")
			})
		}
	})
}

//nolint:thelper
func testCodecWrite[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("write", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				buf := bytes.NewBuffer([]byte{}) // don't let buf.Bytes() return nil
				err := codec.Write(buf, tt.value)
				require.NoError(t, err)
				assert.Equal(t, tt.data, buf.Bytes())
			})
		}
	})
}

//nolint:thelper
func testCodecRead[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				r := bytes.NewReader(tt.data)
				got, err := codec.Read(r)
				require.NoError(t, err)
				assert.Equal(t, 0, r.Len())
				assert.IsType(t, tt.value, got)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("read short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(tt.data)
				if size == 0 {
					return
				}
				// Should either err, or get the wrong value back.
				//nolint:errcheck
				defer func() { recover() }()
				r := bytes.NewReader(tt.data[:size-1])
				got, err := codec.Read(r)
				if err == nil {
					assert.Equal(t, 0, r.Len())
					assert.IsType(t, tt.value, got)
					assert.NotEqual(t, tt.value, got)
				}
			})
		}
	})
}

// Tests input == output, where input => Append/Put/Write => Get/Read => output.
// Does not use testCase.data.
// This is useful when the encoded bytes are indeterminate (unordered maps, e.g.).
//
//nolint:thelper
func testRoundTrip[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer([]byte{})
			err := codec.Write(buf, tt.value)
			require.NoError(t, err)
			got, err := codec.Read(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err)
			assert.IsType(t, tt.value, got)
			assert.Equal(t, tt.value, got)
		})
	}
}

var (
	errRead  = errors.New("failed to read")
	errWrite = errors.New("failed to write")
)

// Tests:
// - Codec.Read() fails when reading from a failing io.Reader.
// - Codec.Write() fails when writing nonEmpty to a failing io.Writer.
func testCodecFail[T any](t *testing.T, codec lexy.Codec[T], nonEmpty T) {
	t.Helper()
	t.Run("fail read", func(t *testing.T) {
		t.Parallel()
		_, err := codec.Read(iotest.ErrReader(errRead))
		assert.Error(t, err)
	})
	t.Run("fail write", func(t *testing.T) {
		t.Parallel()
		err := codec.Write(failWriter{}, nonEmpty)
		assert.Error(t, err)
	})
}

type failWriter struct{}

type boundedWriter struct {
	count, limit int
	data         []byte
}

var (
	_ io.Writer = failWriter{}
	_ io.Writer = &boundedWriter{0, 0, nil}
)

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errWrite
}

// Return number of bytes written from p.
func (w *boundedWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.count
	numToWrite := len(p)
	if numToWrite > remaining {
		numToWrite = remaining
	}
	w.data = append(w.data, p[:numToWrite]...)
	w.count += numToWrite
	if len(p) > remaining {
		return numToWrite, io.EOF
	}
	return numToWrite, nil
}
