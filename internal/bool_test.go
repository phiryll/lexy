package internal

import (
	"bytes"
	"io"
	"testing"
)

func TestBoolCodec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       BoolCodec
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := BoolCodec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("BoolCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BoolCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolCodec_Write(t *testing.T) {
	type args struct {
		value bool
	}
	tests := []struct {
		name    string
		c       BoolCodec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := BoolCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("BoolCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("BoolCodec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
