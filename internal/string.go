package internal

import (
	"io"
	"strings"
)

type StringCodec struct{}

func (c StringCodec) Read(r io.Reader) (string, error) {
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c StringCodec) Write(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}
