package internal

import (
	"fmt"
	"io"
)

type failReader struct{}
type failWriter struct{}

var _ io.Reader = failReader{}
var _ io.Writer = failWriter{}

func (r failReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to read")
}

func (r failWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to write")
}
