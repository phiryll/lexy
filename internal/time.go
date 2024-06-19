package internal

import (
	"io"
	"time"
)

// Needs to encode both the UTC instant and time zone.
// UTC instant: Time.MarshalText() ?
// Time zones should be meaningfully sorted, but how to distinguish daylight-savings from not?

type TimeCodec struct{}

func (c TimeCodec) Read(r io.Reader) (time.Time, error) {
	panic("unimplemented")
}

func (c TimeCodec) Write(w io.Writer, value time.Time) error {
	panic("unimplemented")
}
