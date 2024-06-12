package internal

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

// The same signatures as lexy.Codec, redefined here to avoid a cyclic
// dependency.

type codec[T any] interface {
	Read(r io.Reader) (T, error)
	Write(w io.Writer, value T) error
}

type testCase[T comparable] struct {
	name  string
	value T
	data  []byte
}

// Tests:
// - codec.Read() and codec.Write() for the given test cases
// - codec.Read() fails when reading from an empty []byte
// - codec.Write() fails when given a failing io.Writer
func testCodec[T comparable](t *testing.T, codec codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := codec.Read(bytes.NewReader(tt.data))
				if err != nil {
					t.Errorf("Read() error = %v", err)
					return
				}
				if got != tt.value {
					t.Errorf("Read() = %v, want %v", got, tt.value)
				}
			})
		}
		t.Run("fail", func(t *testing.T) {
			if _, err := codec.Read(bytes.NewReader([]byte{})); err == nil {
				t.Errorf("Read() wantErr")
			}
		})
	})

	t.Run("write", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := &bytes.Buffer{}
				if err := codec.Write(w, tt.value); err != nil {
					t.Errorf("Write() error = %v", err)
					return
				}
				if gotW := w.Bytes(); !bytes.Equal(gotW, tt.data) {
					t.Errorf("Write() = %v, want %v", gotW, tt.data)
				}
			})
		}
		t.Run("fail", func(t *testing.T) {
			var value T
			if err := codec.Write(failWriter{}, value); err == nil {
				t.Errorf("Write() wantErr")
			}
		})
	})
}

type failWriter struct{}

var _ io.Writer = failWriter{}

func (w failWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to write")
}
