// Package lexy defines an API for lexicographically ordered unsigned byte encodings.
//
// TODO: Lots of usage details.
package lexy

import (
	"bytes"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/phiryll/lexy/internal"
)

// Codec defines methods for lexicographically ordered unsigned byte encodings.
//
// Encoded values must have the same order as the values they encode.
// The Read and Write methods should be lossless inverse operations if possible, and clearly documented if not.
//
// Read will read until either it has all the data it needs, or the argument io.Reader stops returning data.
// io.Reader.Read is permitted to return only immediately available data instead of waiting for more.
// This may cause an error (IntCodec32), or it may silently return incomplete data (StringCodec).
//
// Read and Write may have to process data one byte at a time, so using buffered I/O is recommended.
// Never use a buffered Reader wrapping the argument io.Reader within a Codec implementation.
// If you use a buffered Writer within a Codec implementation, it must be flushed before returning.
type Codec[T any] interface {
	// Write writes value to w.
	Write(w io.Writer, value T) error

	// Read reads a value from r and returns it.
	Read(r io.Reader) (T, error)
}

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write if you're encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(&b, value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read if you're decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	// bytes.NewBuffer takes ownership of its argument, so we need to clone it first.
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
