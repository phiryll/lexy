package internal_test

import (
	"bytes"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestBigIntCodec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       internal.BigIntCodec
		args    args
		want    big.Int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.BigIntCodec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("BigIntCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BigIntCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigIntCodec_Write(t *testing.T) {
	type args struct {
		value big.Int
	}
	tests := []struct {
		name    string
		c       internal.BigIntCodec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.BigIntCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("BigIntCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("BigIntCodec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestBigFloatCodec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       internal.BigFloatCodec
		args    args
		want    big.Float
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.BigFloatCodec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("BigFloatCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BigFloatCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigFloatCodec_Write(t *testing.T) {
	type args struct {
		value big.Float
	}
	tests := []struct {
		name    string
		c       internal.BigFloatCodec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.BigFloatCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("BigFloatCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("BigFloatCodec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
