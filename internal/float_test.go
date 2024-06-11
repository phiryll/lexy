package internal

import (
	"bytes"
	"io"
	"testing"
)

func TestFloat32Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Float32Codec
		args    args
		want    float32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Float32Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Float32Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Float32Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat32Codec_Write(t *testing.T) {
	type args struct {
		value float32
	}
	tests := []struct {
		name    string
		c       Float32Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Float32Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Float32Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Float32Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestFloat64Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Float64Codec
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Float64Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Float64Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Float64Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64Codec_Write(t *testing.T) {
	type args struct {
		value float64
	}
	tests := []struct {
		name    string
		c       Float64Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Float64Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Float64Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Float64Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
