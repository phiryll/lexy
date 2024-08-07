package lexy_test

// This file contains things that help in writing Codec tests,
// it doesn't have any tests itself.

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"testing/iotest"
	"time"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Codecs used by tests
var (
	boolCodec  = lexy.Bool[bool]()
	uintCodec  = lexy.Uint[uint]()
	uint8Codec = lexy.Uint8[uint8]()
	// uint16Codec   = lexy.UintCodec[uint16]()
	uint32Codec = lexy.Uint32[uint32]()
	// uint64Codec   = lexy.UintCodec[uint64]()
	intCodec      = lexy.Int[int]()
	int8Codec     = lexy.Int8[int8]()
	int16Codec    = lexy.Int16[int16]()
	int32Codec    = lexy.Int32[int32]()
	int64Codec    = lexy.Int64[int64]()
	float32Codec  = lexy.Float32[float32]()
	float64Codec  = lexy.Float64[float64]()
	durationCodec = lexy.Int64[time.Duration]()
	aStringCodec  = lexy.String[string]()
)

func ptr[T any](value T) *T {
	return &value
}

func encoderFor[T any](codec lexy.Codec[T]) func(value T) []byte {
	return func(value T) []byte {
		data, err := lexy.Encode(codec, value)
		if err != nil {
			panic(err)
		}
		return data
	}
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
// - Codec.Read() and Codec.Write() are invertible for the given test cases
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := codec.Read(bytes.NewReader(tt.data))
				require.NoError(t, err)
				assert.IsType(t, tt.value, got)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("write", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				buf := bytes.NewBuffer([]byte{}) // don't let buf.Bytes() return nil
				err := codec.Write(buf, tt.value)
				require.NoError(t, err)
				assert.Equal(t, tt.data, buf.Bytes())
			})
		}
	})
}

// Tests input == output, where input => Codec.Write => Codec.Read => output.
// Does not use testCase.data.
// This is useful when the encoded bytes are indeterminate (unordered maps and structs, e.g.).
func testCodecRoundTrip[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// Tests:
// - Codec.Read() fails when reading from a failing io.Reader
// - Codec.Write() fails when writing nonEmpty to a failing io.Writer
func testCodecFail[T any](t *testing.T, codec lexy.Codec[T], nonEmpty T) {
	t.Run("fail read", func(t *testing.T) {
		_, err := codec.Read(iotest.ErrReader(fmt.Errorf("failed to read")))
		assert.Error(t, err)
	})
	t.Run("fail write", func(t *testing.T) {
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
	_ io.Writer = &boundedWriter{}
)

func (w failWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to write")
}

// return number written from p
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
