// Package lexy defines an API for lexicographically ordered unsigned byte encodings.
//
// TODO: Lots of usage details.
package lexy

import (
	"io"
	"math/big"
	"time"

	"github.com/phiryll/lexy/internal"
)

// Codec defines methods for lexicographically ordered unsigned byte encodings.
//
// Encoded values should have the same order as the values they encode.
// The Read and Write methods should be lossless inverse operations.
// Exceptions to these berhaviors should be clearly documented.
//
// All Codec implementations in lexy are thread-safe,
// including the Codecs for pointers, slices, and maps if their delegate Codecs are thread-safe.
type Codec[T any] interface {
	// Read will read from r and decode a value of type T.
	//
	// Read will read from r until either it has all the data it needs, or r stops returning data.
	// r.Read is permitted to return only immediately available data instead of waiting for more.
	// This may cause an error, or it may silently return incomplete data, depending on this Reader's imlementation.
	//
	// Read may have to process data one byte at a time, so using a buffered io.Reader is recommended if appropriate.
	// However, never create a buffered io.Reader wrapping the argument io.Reader within a Codec implementation.
	// A buffered io.Reader will read more than necessary to fill its buffer,
	// making any unused bytes unavailable for the next Read, preventing that Codec's use within an aggregate Codec.
	Read(r io.Reader) (T, error)

	// Writer will encode value and write the resulting bytes to w.
	//
	// Write may have to process data one byte at a time, so using a buffered io.Writer is recommended if appropriate.
	// If you use a buffered io.Writer within a Codec implementation, it must be flushed before returning from Write.
	Write(w io.Writer, value T) error

	// RequiresTerminator returns whether this Codec requires a terminator (and therefore escaping)
	// when used within an aggregate Codec (slice, map, or struct).
	RequiresTerminator() bool
}

// Prefixes to use for encodings that would normally encode nil or an empty value as zero bytes.
// The values were chosen so that nil < empty < non-empty, and the prefixes don't need to be escaped.
// This is normally only an issue for variable length encodings.
//
// This prevents ambiguous encodings like these
// (0x00 is the terminator between slice elements, if required):
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
const (
	PrefixNil      byte = internal.PrefixNil
	PrefixEmpty    byte = internal.PrefixEmpty
	PrefixNonEmpty byte = internal.PrefixNonEmpty
)

const (
	// TerminatorByte is used to terminate elements, when necessary.
	TerminatorByte byte = internal.TerminatorByte

	// EscapeByte is used the escape the terminator and escape bytes when they appear in data, when necessary.
	// This includes those values appearing in the encodings of nested aggregates,
	// because those are still just data at the level of the enclosing aggregate.
	EscapeByte byte = internal.EscapeByte
)

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write if you're encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	return internal.Encode(codec, value)
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read if you're decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return internal.Decode(codec, data)
}

// Codecs that do not delegate to other Codecs, for types with builtin underlying types.

func BoolCodec[T ~bool]() Codec[T]                                { return internal.UintCodec[T]() }
func UIntCodec[T ~uint8 | ~uint16 | ~uint32 | ~uint64]() Codec[T] { return internal.UintCodec[T]() }
func IntCodec[T ~int8 | ~int16 | ~int32 | ~int64]() Codec[T]      { return internal.IntCodec[T]() }
func Float32Codec[T ~float32]() Codec[T]                          { return internal.Float32Codec[T]() }
func Float64Codec[T ~float64]() Codec[T]                          { return internal.Float64Codec[T]() }
func Complex64Codec() Codec[complex64]                            { return internal.Complex64Codec }
func Complex128Codec() Codec[complex128]                          { return internal.Complex128Codec }
func StringCodec[T ~string]() Codec[T]                            { return internal.StringCodec[T]() }
func DurationCodec() Codec[time.Duration]                         { return internal.IntCodec[time.Duration]() }

// Codecs that do not delegate to other Codecs, for types without builtin underlying types (all structs).

func BigIntCodec() Codec[*big.Int]     { return internal.BigIntCodec }
func BigFloatCodec() Codec[*big.Float] { return internal.BigFloatCodec }
func TimeCodec() Codec[time.Time]      { return internal.TimeCodec }

// Codecs that delegate to other Codecs.

// NegateCodec returns a new Codec reversing the encoding order produced by codec.
func NegateCodec[T any](codec Codec[T]) Codec[T] {
	return internal.MakeNegateCodec(codec)
}

func PointerCodec[P ~*T, T any](elemCodec Codec[T]) Codec[P] {
	return internal.MakePointerCodec[P](elemCodec)
}

func SliceCodec[S ~[]T, T any](elemCodec Codec[T]) Codec[S] {
	return internal.MakeSliceCodec[S](elemCodec)
}

func MapCodec[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return internal.MakeMapCodec[M](keyCodec, valueCodec)
}
