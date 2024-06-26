package internal

import (
	"io"
)

// Same interface as lexy.Codec, to avoid a circular dependency.
type codec[T any] interface {
	Write(w io.Writer, value T) error
	Read(r io.Reader) (T, error)
}
