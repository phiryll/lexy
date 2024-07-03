package internal

import (
	"io"
	"time"
)

var (
	TimeCodec codec[time.Time] = timeCodec{}
)

// Needs to encode both the UTC instant and time zone.
// UTC instant: Time.MarshalText() ?
// Time zones should be meaningfully sorted, but how to distinguish daylight-savings from not?

type timeCodec struct{}

func (c timeCodec) Read(r io.Reader) (time.Time, error) {
	panic("unimplemented")
}

func (c timeCodec) Write(w io.Writer, value time.Time) error {
	panic("unimplemented")
}
