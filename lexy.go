package lexy

import (
	"bytes"
	"io"
	"math/big"
	"time"

	"github.com/phiryll/lexy/internal"
)

// Codec defines methods for encoding and decoding values to and from a binary
// form.
type Codec[T any] interface {
	// Unfortunately, a Codec can't be defined or created using
	// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler. Those types
	// require the value to be a receiver instead of an argument.

	// Write writes a value to the given io.Writer.
	Write(value T, w io.Writer) error

	// Read reads a value from the given io.Reader and returns it.
	Read(r io.Reader) (T, error)
}

// Encode uses codec to encode value into a []byte and returns it. This is a
// convenience function to create a new []byte.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(value, &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode uses codec to decode data into a value and returns it. This is a
// convenience function using a []byte.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	// TODO: NewBuffer takes ownership of data, this is probably a bad idea
	// here.
	return codec.Read(bytes.NewBuffer(data))
}

func BoolCodec() Codec[bool]                        { return internal.BoolCodec{} }
func UInt8Codec() Codec[uint8]                      { return internal.Uint8Codec{} }
func UInt16Codec() Codec[uint16]                    { return internal.Uint16Codec{} }
func UInt32Codec() Codec[uint32]                    { return internal.Uint32Codec{} }
func UInt64Codec() Codec[uint64]                    { return internal.Uint64Codec{} }
func Int8Codec() Codec[int8]                        { return internal.Int8Codec{} }
func Int16Codec() Codec[int16]                      { return internal.Int16Codec{} }
func Int32Codec() Codec[int32]                      { return internal.Int32Codec{} }
func Int64Codec() Codec[int64]                      { return internal.Int64Codec{} }
func Float32Codec() Codec[float32]                  { return internal.Float32Codec{} }
func Float64Codec() Codec[float64]                  { return internal.Float64Codec{} }
func BigIntCodec() Codec[big.Int]                   { return internal.BigIntCodec{} }
func BigFloatCodec() Codec[big.Float]               { return internal.BigFloatCodec{} }
func StringCodec() Codec[string]                    { return internal.StringCodec{} }
func TimeCodec() Codec[time.Time]                   { return internal.TimeCodec{} }
func SliceCodec[T any]() Codec[[]T]                 { return internal.SliceCodec[T]{} }
func MapCodec[K comparable, V any]() Codec[map[K]V] { return internal.MapCodec[K, V]{} }
func StructCodec[T any]() Codec[T]                  { return internal.StructCodec[T]{} }

// Decouple type prefixes, those are a feature of the aggregate Codec. The int8,
// int16, ... Codecs should not have prefixes.
