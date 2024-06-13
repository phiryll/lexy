package lexy

import (
	"bytes"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/phiryll/lexy/internal"
)

// Codec defines methods for encoding and decoding values to and from a
// binary representation. In Lexy, the lexicographical unsigned byte
// order of the representations of two values should have the same order
// as the unencoded values. A Codec's Read and Write methods should be
// lossless inverse operations if possible, and clearly documented if
// not.
type Codec[T any] interface {
	// Unfortunately, a Codec can't be defined or created using
	// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler. Those
	// types require the value to be a receiver instead of an argument.

	// Read reads a value from the given io.Reader and returns it.
	Read(r io.Reader) (T, error)

	// Write writes a value to the given io.Writer.
	Write(w io.Writer, value T) error
}

// Encode uses codec to encode value into a []byte and returns it. This
// is a convenience function to create a new []byte.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(&b, value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode uses codec to decode data into a value and returns it. This is
// a convenience function using a []byte.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	// bytes.NewBuffer takes ownership of its argument, so we need to
	// clone it first.
	return codec.Read(bytes.NewBuffer(bytes.Clone(data)))
}

func BoolCodec() Codec[bool]                        { return internal.UintCodec[bool]{} }
func UInt8Codec() Codec[uint8]                      { return internal.UintCodec[uint8]{} }
func UInt16Codec() Codec[uint16]                    { return internal.UintCodec[uint16]{} }
func UInt32Codec() Codec[uint32]                    { return internal.UintCodec[uint32]{} }
func UInt64Codec() Codec[uint64]                    { return internal.UintCodec[uint64]{} }
func Int8Codec() Codec[int8]                        { return internal.IntCodec[int8]{Mask: math.MinInt8} }
func Int16Codec() Codec[int16]                      { return internal.IntCodec[int16]{Mask: math.MinInt16} }
func Int32Codec() Codec[int32]                      { return internal.IntCodec[int32]{Mask: math.MinInt32} }
func Int64Codec() Codec[int64]                      { return internal.IntCodec[int64]{Mask: math.MinInt64} }
func Float32Codec() Codec[float32]                  { return internal.Float32Codec{} }
func Float64Codec() Codec[float64]                  { return internal.Float64Codec{} }
func BigIntCodec() Codec[big.Int]                   { return internal.BigIntCodec{} }
func BigFloatCodec() Codec[big.Float]               { return internal.BigFloatCodec{} }
func StringCodec() Codec[string]                    { return internal.StringCodec{} }
func TimeCodec() Codec[time.Time]                   { return internal.TimeCodec{} }
func SliceCodec[T any]() Codec[[]T]                 { return internal.SliceCodec[T]{} }
func MapCodec[K comparable, V any]() Codec[map[K]V] { return internal.MapCodec[K, V]{} }
func StructCodec[T any]() Codec[T]                  { return internal.StructCodec[T]{} }

// TODO: expose escaping functions, and document them well!
