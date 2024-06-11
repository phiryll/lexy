package internal

import (
	"bytes"
	"io"
	"testing"
)

func TestStringCodec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       StringCodec
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := StringCodec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringCodec_Write(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		c       StringCodec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := StringCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("StringCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("StringCodec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
