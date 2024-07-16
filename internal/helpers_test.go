package internal_test

// This file contains things that help in writing Codec tests,
// it doesn't have any tests itself.

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/phiryll/lexy"
	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Codecs used by tests
var (
	boolCodec     = internal.UintCodec[bool]()
	uint8Codec    = internal.UintCodec[uint8]()
	uint16Codec   = internal.UintCodec[uint16]()
	uint32Codec   = internal.UintCodec[uint32]()
	uint64Codec   = internal.UintCodec[uint64]()
	int8Codec     = internal.IntCodec[int8]()
	int16Codec    = internal.IntCodec[int16]()
	int32Codec    = internal.IntCodec[int32]()
	int64Codec    = internal.IntCodec[int64]()
	float32Codec  = internal.Float32Codec[float32]()
	float64Codec  = internal.Float64Codec[float64]()
	bigIntCodec   = internal.BigIntCodec
	bigFloatCodec = internal.BigFloatCodec
	stringCodec   = internal.StringCodec[string]()
	timeCodec     = internal.TimeCodec
	durationCodec = internal.IntCodec[time.Duration]()
)

func ptr[T any](value T) *T {
	return &value
}

type testCase[T any] struct {
	name  string
	value T
	data  []byte
}

// Just to make the test cases terser.
const (
	del      byte = internal.DelimiterByte
	esc      byte = internal.EscapeByte
	empty    byte = internal.PrefixEmpty
	nonEmpty byte = internal.PrefixNonEmpty
)

// Tests:
// - Codec.Read() and Codec.Write() are invertible for the given test cases
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := codec.Read(bytes.NewReader(tt.data))
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("write", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var b bytes.Buffer
				err := codec.Write(&b, tt.value)
				require.NoError(t, err)
				assert.Equal(t, tt.data, b.Bytes())
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
			var b bytes.Buffer
			err := codec.Write(&b, tt.value)
			require.NoError(t, err)
			got, err := codec.Read(bytes.NewReader(b.Bytes()))
			require.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

// Tests:
// - Codec.Read() fails when reading from a failing io.Reader
// - Codec.Write() fails when writing nonEmpty to a failing io.Writer
func testCodecFail[T any](t *testing.T, codec lexy.Codec[T], nonEmpty T) {
	t.Run("fail read", func(t *testing.T) {
		_, err := codec.Read(failReader{})
		assert.Error(t, err)
	})
	t.Run("fail write", func(t *testing.T) {
		err := codec.Write(failWriter{}, nonEmpty)
		assert.Error(t, err)
	})
}

type failReader struct{}
type failWriter struct{}

type boundedWriter struct {
	count, limit int
	data         []byte
}

var (
	_ io.Reader = failReader{}
	_ io.Writer = failWriter{}
	_ io.Writer = &boundedWriter{}
)

func (f failReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to read")
}

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
