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
// All Codecs provided by lexy are safe for concurrent use if their delegate Codecs (if any) are,
// except for Codecs created by Terminate and TerminateIfNeeded.
type Codec[T any] interface {
	// Read will read from r and decode a value of type T.
	//
	// Read will read from r until either it has all the data it needs, or r stops returning data.
	// r.Read is permitted to return only immediately available data instead of waiting for more.
	// This may cause an error, or it may silently return incomplete data, depending on this Reader's implementation.
	//
	// Read may have to process data one byte at a time, so using a buffered io.Reader is recommended if appropriate.
	// However, never create a buffered io.Reader wrapping the argument io.Reader within a Codec implementation.
	// A buffered io.Reader will read more than necessary to fill its buffer,
	// making any unused bytes unavailable for the next Read, preventing that Codec's use within an aggregate Codec.
	Read(r io.Reader) (T, error)

	// Writer will encode value and write the encoded bytes to w.
	//
	// Write may have to process data one byte at a time, so using a buffered io.Writer is recommended if appropriate.
	// If a buffered io.Writer is used within a Codec implementation, it must be flushed before returning from Write.
	Write(w io.Writer, value T) error

	// RequiresTerminator returns whether this Codec requires a terminator (and therefore escaping)
	// when used within an aggregate Codec (for example, within a slice, map, or struct).
	RequiresTerminator() bool
}

// Codecs that do not delegate to other Codecs, for types with builtin underlying types.

func Bool[T ~bool]() Codec[T]                                { return internal.UintCodec[T]() }
func Uint[T ~uint8 | ~uint16 | ~uint32 | ~uint64]() Codec[T] { return internal.UintCodec[T]() }
func AsUint64[T ~uint]() Codec[T]                            { return internal.AsUint64Codec[T]() }
func Int[T ~int8 | ~int16 | ~int32 | ~int64]() Codec[T]      { return internal.IntCodec[T]() }
func AsInt64[T ~int]() Codec[T]                              { return internal.AsInt64Codec[T]() }
func Float32[T ~float32]() Codec[T]                          { return internal.Float32Codec[T]() }
func Float64[T ~float64]() Codec[T]                          { return internal.Float64Codec[T]() }
func Complex64() Codec[complex64]                            { return internal.Complex64Codec }
func Complex128() Codec[complex128]                          { return internal.Complex128Codec }
func String[T ~string]() Codec[T]                            { return internal.StringCodec[T]() }
func Duration() Codec[time.Duration]                         { return internal.IntCodec[time.Duration]() }

// Codecs that do not delegate to other Codecs, for types without builtin underlying types (all structs).

func Time() Codec[time.Time]              { return internal.TimeCodec }
func BigInt() Codec[*big.Int]             { return internal.BigIntCodec(true) }
func BigIntNilsLast() Codec[*big.Int]     { return internal.BigIntCodec(false) }
func BigFloat() Codec[*big.Float]         { return internal.BigFloatCodec(true) }
func BigFloatNilsLast() Codec[*big.Float] { return internal.BigFloatCodec(false) }
func BigRat() Codec[*big.Rat]             { return internal.BigRatCodec(true) }
func BigRatNilsLast() Codec[*big.Rat]     { return internal.BigRatCodec(false) }

// Codecs that delegate to other Codecs.

func PointerTo[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerCodec[P](elemCodec, true)
}

func PointerToNilsLast[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerCodec[P](elemCodec, false)
}

func ArrayOf[A any, E any](elemCodec Codec[E]) Codec[A] {
	return internal.ArrayCodec[A](elemCodec)
}

func PointerToArrayOf[P ~*A, A any, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerToArrayCodec[P](elemCodec, true)
}

func PointerToArrayOfNilsLast[P ~*A, A any, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerToArrayCodec[P](elemCodec, false)
}

func SliceOf[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	return internal.SliceCodec[S](elemCodec, true)
}

func SliceOfNilsLast[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	return internal.SliceCodec[S](elemCodec, false)
}

func Bytes[S ~[]byte]() Codec[S] {
	return internal.BytesCodec[S](true)
}

func BytesNilsLast[S ~[]byte]() Codec[S] {
	return internal.BytesCodec[S](false)
}

func MapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return internal.MapCodec[M](keyCodec, valueCodec, true)
}

func MapOfNilsLast[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return internal.MapCodec[M](keyCodec, valueCodec, false)
}

// Negate returns a new Codec reversing the encoding order produced by codec.
func Negate[T any](codec Codec[T]) Codec[T] {
	return internal.NegateCodec(codec)
}

func Terminate[T any](codec Codec[T]) Codec[T] {
	return internal.Terminate(codec)
}

func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	return internal.TerminateIfNeeded(codec)
}

func ReadPrefix[T any](r io.Reader, nilable bool, emptyValue *T) (value T, done bool, err error) {
	return internal.ReadPrefix(r, nilable, emptyValue)
}

func WritePrefixNilsFirst[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	return internal.WritePrefixNilsFirst(w, isNil, isEmpty, value)
}

func WritePrefixNilsLast[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	return internal.WritePrefixNilsLast(w, isNil, isEmpty, value)
}

func IsNilPointer[P ~*E, E any](value P) bool {
	return value == nil
}

func IsNilSlice[S ~[]E, E any](value S) bool {
	return value == nil
}

func IsNilMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value == nil
}

func IsEmptyString[T ~string](value T) bool {
	return len(value) == 0
}

func IsEmptySlice[S ~[]E, E any](value S) bool {
	return value != nil && len(value) == 0
}

func IsEmptyMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value != nil && len(value) == 0
}

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write when encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	return internal.Encode(codec, value)
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read when decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return internal.Decode(codec, data)
}
