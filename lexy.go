package lexy

import (
	"bytes"
	"io"
)

var UINT8Codec Codec[uint8] = uint8Codec{}

// Codec defines methods for encoding and decoding values to and from
// a binary form.
type Codec[T any] interface {
	// Unfortunately, a Codec can't be defined or created using
	// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler.
	// Those types require the value to be a receiver instead of
	// an argument.

	// Write writes a value to the given io.Writer.
	Write(value T, w io.Writer) error

	// Read reads a value from the given io.Reader and returns it.
	Read(r io.Reader) (T, error)
}

// Encode uses codec to encode value into a []byte and returns it.
// This is a convenience function for when the []byte should only
// contain one encoded value.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(value, &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode decodes the given binary form into a value and returns it.
// Decode must be able to decode the form generated by Encode. This is
// a convenience function used when the []byte only contains one
// value.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return codec.Read(bytes.NewBuffer(data))
}

type uint8Codec struct{}

func (uint8Codec) Write(value uint8, w io.Writer) error {
	// TODO
	return nil
}

func (uint8Codec) Read(r io.Reader) (uint8, error) {
	// TODO
	return 0, nil
}

// Completely decouple type prefixes, those are a feature of the
// aggregate Codec. The int8, int16, ... Codecs should not have
// prefixes.
