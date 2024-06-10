package internal

import (
	"bytes"
	"io"
	"testing"
)

func TestUint8Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Uint8Codec
		args    args
		want    uint8
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint8Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint8Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint8Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint8Codec_Write(t *testing.T) {
	type args struct {
		value uint8
	}
	tests := []struct {
		name    string
		c       Uint8Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint8Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Uint8Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Uint8Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestUint16Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Uint16Codec
		args    args
		want    uint16
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint16Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint16Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint16Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint16Codec_Write(t *testing.T) {
	type args struct {
		value uint16
	}
	tests := []struct {
		name    string
		c       Uint16Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint16Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Uint16Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Uint16Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestUint32Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Uint32Codec
		args    args
		want    uint32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint32Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint32Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint32Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint32Codec_Write(t *testing.T) {
	type args struct {
		value uint32
	}
	tests := []struct {
		name    string
		c       Uint32Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint32Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Uint32Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Uint32Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestUint64Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Uint64Codec
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint64Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint64Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint64Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint64Codec_Write(t *testing.T) {
	type args struct {
		value uint64
	}
	tests := []struct {
		name    string
		c       Uint64Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Uint64Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Uint64Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Uint64Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestInt8Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Int8Codec
		args    args
		want    int8
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int8Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int8Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int8Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt8Codec_Write(t *testing.T) {
	type args struct {
		value int8
	}
	tests := []struct {
		name    string
		c       Int8Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int8Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Int8Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Int8Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestInt16Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Int16Codec
		args    args
		want    int16
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int16Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int16Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int16Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt16Codec_Write(t *testing.T) {
	type args struct {
		value int16
	}
	tests := []struct {
		name    string
		c       Int16Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int16Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Int16Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Int16Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestInt32Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Int32Codec
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int32Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int32Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int32Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt32Codec_Write(t *testing.T) {
	type args struct {
		value int32
	}
	tests := []struct {
		name    string
		c       Int32Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int32Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Int32Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Int32Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestInt64Codec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       Int64Codec
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int64Codec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64Codec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int64Codec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64Codec_Write(t *testing.T) {
	type args struct {
		value int64
	}
	tests := []struct {
		name    string
		c       Int64Codec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Int64Codec{}
			w := &bytes.Buffer{}
			if err := c.Write(tt.args.value, w); (err != nil) != tt.wantErr {
				t.Errorf("Int64Codec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Int64Codec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
