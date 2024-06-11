package internal

import (
	"io"
	"time"
)

type TimeCodec struct{}

func (c TimeCodec) Read(r io.Reader) (time.Time, error) {
	panic("unimplemented")
}

func (c TimeCodec) Write(w io.Writer, value time.Time) error {
	panic("unimplemented")
}
