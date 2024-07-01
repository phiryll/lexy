package internal

import (
	"io"
)

// Same interface as lexy.Codec, to avoid a circular dependency.
// lexy.Codec cannot be a type alias to this, because generic type aliases are not permitted.
type codec[T any] interface {
	Write(w io.Writer, value T) error
	Read(r io.Reader) (T, error)
}

// Prefixes, documented in lexy.go
const (
	// 0x02 is reserved for nil if that becomes necessary.
	PrefixZero    byte = 0x03
	PrefixNonZero byte = 0x04
)

// Convenience byte slices.
var (
	prefixZero    = []byte{PrefixZero}
	prefixNonZero = []byte{PrefixNonZero}
)
