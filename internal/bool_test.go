package internal

import (
	"bytes"
	"io"
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
			c := BoolCodec{}
			got, err := c.Read(tt.r)
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
			c := BoolCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.value, w); (err != nil) != tt.wantErr {
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
	c := BoolCodec{}
	w := failWriter{}
	if err := c.Write(true, w); err == nil {
		t.Errorf("BoolCodec.Write() error = %v, wantErr %v", err, true)
	}
}
