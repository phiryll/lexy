package internal

import (
	"bytes"
	"io"
	"math"
	"testing"
)

func TestBoolCodec_Read(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    bool
		wantErr bool
	}{
		{"false", bytes.NewReader([]byte{0}), false, false},
		{"true", bytes.NewReader([]byte{1}), true, false},
		{"fail", bytes.NewReader([]byte{}), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UintCodec[bool]{}
			got, err := c.Read(tt.r)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("BoolCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("BoolCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolCodec_Write(t *testing.T) {
	tests := []struct {
		name    string
		value   bool
		want    []byte
		wantErr bool
	}{
		{"false", false, []byte{0}, false},
		{"true", true, []byte{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UintCodec[bool]{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("BoolCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.Bytes(); !bytes.Equal(gotW, tt.want) {
				t.Errorf("BoolCodec.Write() = %v, want %v", gotW, tt.want)
			}
		})
	}
}

func TestBoolCodec_WriteFail(t *testing.T) {
	c := UintCodec[bool]{}
	w := failWriter{}
	if err := c.Write(w, true); err == nil {
		t.Errorf("BoolCodec.Write() error = %v, wantErr %v", err, true)
	}
}

func TestUint8Codec_Read(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    uint8
		wantErr bool
	}{
		{"0x00", bytes.NewReader([]byte{0x00}), 0x00, false},
		{"0x01", bytes.NewReader([]byte{0x01}), 0x01, false},
		{"0x7F", bytes.NewReader([]byte{0x7F}), 0x7F, false},
		{"0x80", bytes.NewReader([]byte{0x80}), 0x80, false},
		{"0xFF", bytes.NewReader([]byte{0xFF}), 0xFF, false},
		{"fail", bytes.NewReader([]byte{}), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UintCodec[uint8]{}
			got, err := c.Read(tt.r)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Uint8Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("Uint8Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint8Codec_Write(t *testing.T) {
	tests := []struct {
		name    string
		value   uint8
		wantW   []byte
		wantErr bool
	}{
		{"0x00", 0x00, []byte{0x00}, false},
		{"0x01", 0x01, []byte{0x01}, false},
		{"0x7F", 0x7F, []byte{0x7F}, false},
		{"0x80", 0x80, []byte{0x80}, false},
		{"0xFF", 0xFF, []byte{0xFF}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UintCodec[uint8]{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Uint8Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.Bytes(); !bytes.Equal(gotW, tt.wantW) {
				t.Errorf("Uint8Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestInt8Codec_Read(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		want    int8
		wantErr bool
	}{
		{"min", bytes.NewReader([]byte{0x00}), math.MinInt8, false},
		{"-1", bytes.NewReader([]byte{0x7F}), -1, false},
		{"0", bytes.NewReader([]byte{0x80}), 0, false},
		{"+1", bytes.NewReader([]byte{0x81}), 1, false},
		{"max", bytes.NewReader([]byte{0xFF}), math.MaxInt8, false},
		{"fail", bytes.NewReader([]byte{}), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := IntCodec[int8]{Mask: math.MinInt8}
			got, err := c.Read(tt.r)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Int8Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("Int8Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt8Codec_Write(t *testing.T) {
	tests := []struct {
		name    string
		value   int8
		wantW   []byte
		wantErr bool
	}{
		{"min", math.MinInt8, []byte{0x00}, false},
		{"-1", -1, []byte{0x7F}, false},
		{"0", 0, []byte{0x80}, false},
		{"+1", 1, []byte{0x81}, false},
		{"max", math.MaxInt8, []byte{0xFF}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := IntCodec[int8]{Mask: math.MinInt8}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Int8Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.Bytes(); !bytes.Equal(gotW, tt.wantW) {
				t.Errorf("Int8Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
