// Package lexy defines an API for lexicographically ordered unsigned byte encodings.
//
// TODO: Lots of usage details.
package lexy

import (
	"bytes"
	"io"
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
// This may cause an error (Int32Codec), or it may silently return incomplete data (StringCodec).
//
// Read and Write may have to process data one byte at a time, so using buffered I/O is recommended.
// Never use a buffered Reader wrapping the argument io.Reader within a Codec implementation.
// If you use a buffered Writer within a Codec implementation, it must be flushed before returning.
//
// All Codec implementations in lexy are thread-safe,
// including the codecs for slices, maps, and structs if their delegate Codecs are thread-safe.
type Codec[T any] interface {
	// Write writes value to w.
	Write(w io.Writer, value T) error

	// Read reads a value from r and returns it.
	Read(r io.Reader) (T, error)
}

// Prefixes to use for encodings that would normally encode an empty value as zero bytes.
// Only nil should be encoded as zero bytes, empty is not nil in lexy for all types.
// The values were chosen so that nil < empty < non-empty, and the prefixes don't need to be escaped.
// This is normally only an issue for variable length encodings.
//
// This prevents ambiguous encodings like these (0x00 is the delimiter between slice elements):
//
//	""                     => []
//
//	[]string{}             => []
//	[]string{""}           => []
//
//	[][]string{{}, {}}     => [0x00]
//	[][]string{{}, {""}}   => [0x00]
//	[][]string{{""}, {}}   => [0x00]
//	[][]string{{""}, {""}} => [0x00]
//
// which would instead be encoded as (in sort order within groups):
//
//	""                     => [0x03]
//	                          [empty-string]
//
//	[]string{}             => [0x03]
//	                          [empty-slice]
//	[]string{""}           => [0x04, 0x03]
//	                          [non-empty-slice, empty-string]
//
//	[][]string{{}, {}}     => [0x04, 0x03, 0x00, 0x03]
//	                          [non-empty-slice,
//	                             empty-slice, delim,
//	                             empty-slice]
//	[][]string{{}, {""}}   => [0x04, 0x03, 0x00, 0x04, 0x03]
//	                          [non-empty-slice,
//	                             empty-slice, delim,
//	                             non-empty-slice, empty-string]
//	[][]string{{""}, {}}   => [0x04, 0x04, 0x03, 0x00, 0x03]
//	                          [non-empty-slice,
//	                             non-empty-slice, empty-string, delim,
//	                             empty-slice]
//	[][]string{{""}, {""}} => [0x04, 0x04, 0x03, 0x00, 0x04, 0x03]
//	                          [non-empty-slice,
//	                             non-empty-slice, empty-string, delim,
//	                             non-empty-slice, empty-string]
const (
	// 0x02 is reserved for nil if that becomes necessary.
	PrefixEmpty    byte = internal.PrefixEmpty
	PrefixNonEmpty byte = internal.PrefixNonEmpty
)

const (
	// DelimiterByte is used to delimit elements of an aggregate value.
	DelimiterByte byte = internal.DelimiterByte

	// EscapeByte is used the escape the delimiter and escape bytes when they appear in data.
	//
	// This includes appearing in the encodings of nested aggregates,
	// because those are still just data at the level of the enclosing aggregate.
	EscapeByte byte = internal.EscapeByte
)

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

// Codecs that only need one instance, and have only primitive fields.

func BoolCodec() Codec[bool]          { return internal.BoolCodec }
func UInt8Codec() Codec[uint8]        { return internal.Uint8Codec }
func UInt16Codec() Codec[uint16]      { return internal.Uint16Codec }
func UInt32Codec() Codec[uint32]      { return internal.Uint32Codec }
func UInt64Codec() Codec[uint64]      { return internal.Uint64Codec }
func Int8Codec() Codec[int8]          { return internal.Int8Codec }
func Int16Codec() Codec[int16]        { return internal.Int16Codec }
func Int32Codec() Codec[int32]        { return internal.Int32Codec }
func Int64Codec() Codec[int64]        { return internal.Int64Codec }
func Float32Codec() Codec[float32]    { return internal.Float32Codec }
func Float64Codec() Codec[float64]    { return internal.Float64Codec }
func BigIntCodec() Codec[big.Int]     { return internal.BigIntCodec }
func BigFloatCodec() Codec[big.Float] { return internal.BigFloatCodec }
func StringCodec() Codec[string]      { return internal.StringCodec }
func TimeCodec() Codec[time.Time]     { return internal.TimeCodec }

// Codecs that delegate to other Codecs.

func SliceCodec[T any](elementCodec Codec[T]) Codec[[]T] {
	return internal.NewSliceCodec[T](elementCodec)
}

func MapCodec[K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[map[K]V] {
	return internal.NewMapCodec[K, V](keyCodec, valueCodec)
}

// The entries in the encoded map are ordered by the encodings of its keys.
// The map created by Codec.Read is a normal map, it is not ordered.
func OrderedMapCodec[K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[map[K]V] {
	return internal.NewOrderedMapCodec[K, V](keyCodec, valueCodec)
}

func StructCodec[T any, F any](fieldCodec Codec[F]) Codec[T] {
	// TBD
	return internal.NewStructCodec[T, F](fieldCodec)
}
