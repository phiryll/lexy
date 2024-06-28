package internal_test

// This file contains things that help in writing Codec tests,
// it doesn't have any tests itself.

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase[T any] struct {
	name  string
	value T
	data  []byte
}

// Tests:
// - codec.Read() and codec.Write() are invertible for the given test cases
// - codec.Read() fails when reading from a failing io.Reader
// - codec.Write() fails when writing the zero value of T to a failing io.Writer
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := codec.Read(bytes.NewReader(tt.data))
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			})
		}
		t.Run("fail", func(t *testing.T) {
			_, err := codec.Read(failReader{})
			assert.Error(t, err)
		})
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
		t.Run("fail", func(t *testing.T) {
			var zero T
			err := codec.Write(failWriter{}, zero)
			require.Error(t, err)
		})
	})
}

type failReader struct{}
type failWriter struct{}

type boundedWriter struct {
	count, limit int
	data         []byte
}

var _ io.Reader = failReader{}
var _ io.Writer = failWriter{}
var _ io.Writer = &boundedWriter{}

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
