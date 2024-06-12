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

type readTestCase[T comparable] struct {
	name    string
	data    []byte
	want    T
	wantErr bool
}

func testRead[T comparable](t *testing.T, codec codec[T], tests []readTestCase[T]) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.Read(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

type writeTestCase[T any] struct {
	name    string
	w       byteWriter
	value   T
	want    []byte
	wantErr bool
}

func testWrite[T any](t *testing.T, codec codec[T], tests []writeTestCase[T]) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := codec.Write(tt.w, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := tt.w.Bytes(); !bytes.Equal(gotW, tt.want) {
				t.Errorf("Write() = %v, want %v", gotW, tt.want)
			}
		})
	}
}

// TODO: remove failReader if it ends up being unused.

// Unit test are slightly simpler if failWriter also implements a
// function with the same signature as bytes.Buffer.Bytes().
type byteWriter interface {
	io.Writer
	Bytes() []byte
}

type failReader struct{}
type failWriter struct{}

var _ io.Reader = failReader{}
var _ byteWriter = failWriter{}

func (r failReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to read")
}

func (w failWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to write")
}

func (w failWriter) Bytes() []byte {
	return nil
}
