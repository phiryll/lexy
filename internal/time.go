package internal

import (
	"io"
	"time"
)

type TimeCodec struct{}

func (c TimeCodec) Read(r io.Reader) (time.Time, error) {
	panic("unimplemented")
}

func (c TimeCodec) Write(value time.Time, w io.Writer) error {
	panic("unimplemented")
}
