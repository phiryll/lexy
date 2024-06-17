package internal_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/phiryll/lexy/internal"
)

func TestTimeCodec_Read(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		c       internal.TimeCodec
		args    args
		want    time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.TimeCodec{}
			got, err := c.Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeCodec.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TimeCodec.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeCodec_Write(t *testing.T) {
	type args struct {
		value time.Time
	}
	tests := []struct {
		name    string
		c       internal.TimeCodec
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := internal.TimeCodec{}
			w := &bytes.Buffer{}
			if err := c.Write(w, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("TimeCodec.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("TimeCodec.Write() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
