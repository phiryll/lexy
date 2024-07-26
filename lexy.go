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
	// when used within an aggregate Codec (for example, within a slice, map, or struct Codec).
	// RequiresTerminator should return true if Read does not know when to stop reading,
	// which is the case for unbounded types like slices and maps.
	// If true and used within an aggregate Codec, the aggregate Codec should wrap this Codec
	// with Terminate or TerminateIfNeeded.
	// The wrapping Codec will then limit Read by only reading up to the terminator.
	RequiresTerminator() bool
}

// Codecs that do not delegate to other Codecs, for types with builtin underlying types.

// Bool creates a new Codec for a type with an underlying type of bool.
// This Codec does not require a terminator when used within an aggregate Codec.
func Bool[T ~bool]() Codec[T] { return internal.UintCodec[T]() }

// Uint creates a new Codec for a type with an underlying type of uint8, uint16, uint32, or uint64.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint[T ~uint8 | ~uint16 | ~uint32 | ~uint64]() Codec[T] { return internal.UintCodec[T]() }

// AsUint64 creates a new Codec for a type with an underlying type of uint.
// Values are converted to/from uint64 and encoded with Uint[uint64]().
// This Codec does not require a terminator when used within an aggregate Codec.
func AsUint64[T ~uint]() Codec[T] { return internal.AsUint64Codec[T]() }

// Int creates a new Codec for a type with an underlying type of int8, int16, int32, or int64.
// This Codec does not require a terminator when used within an aggregate Codec.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int[T ~int8 | ~int16 | ~int32 | ~int64]() Codec[T] { return internal.IntCodec[T]() }

// AsInt64 creates a new Codec for a type with an underlying type of int.
// Values are converted to/from int64 and encoded with Int[int64]().
// This Codec does not require a terminator when used within an aggregate Codec.
// This Codec does not require a terminator when used within an aggregate Codec.
func AsInt64[T ~int]() Codec[T] { return internal.AsInt64Codec[T]() }

// Float32 creates a new Codec for a type with an underlying type of float32.
// All bits of the value are preserved by this encoding; NaN values are not canonicalized.
// The encodings for NaNs are merely bytes and are therefore comparable, unlike float32 NaNs.
// There are many different bit patterns for NaN, and their encodings will be distinct.
// No distinction is made between quiet and signaling NaNs.
// This Codec does not require a terminator when used within an aggregate Codec.
// The order of encoded values is:
//
//	-NaN
//	-Inf
//	negative finite numbers
//	-0.0
//	+0.0
//	positive finite numbers
//	+Inf
//	+NaN
func Float32[T ~float32]() Codec[T] { return internal.Float32Codec[T]() }

// Float64 creates a new Codec for a type with an underlying type of float64.
// All bits of the value are preserved by this encoding; NaN values are not canonicalized.
// The encodings for NaNs are merely bytes and are therefore comparable, unlike float64 NaNs.
// There are many different bit patterns for NaN, and their encodings will be distinct.
// No distinction is made between quiet and signaling NaNs.
// This Codec does not require a terminator when used within an aggregate Codec.
// The order of encoded values is:
//
//	-NaN
//	-Inf
//	negative finite numbers
//	-0.0
//	+0.0
//	positive finite numbers
//	+Inf
//	+NaN
func Float64[T ~float64]() Codec[T] { return internal.Float64Codec[T]() }

// Complex64 returns the Codec for the complex64 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for Float32.
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex64() Codec[complex64] { return internal.Complex64Codec }

// Complex128 returns the Codec for the complex128 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for Float64.
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex128() Codec[complex128] { return internal.Complex128Codec }

// String creates a new Codec for a type with an underlying type of string.
// This Codec requires a terminator when used within an aggregate Codec.
func String[T ~string]() Codec[T] { return internal.StringCodec[T]() }

// Duration creates a new Codec for the time.Duration type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Duration() Codec[time.Duration] { return internal.IntCodec[time.Duration]() }

// Codecs that do not delegate to other Codecs, for types without builtin underlying types (all structs).

// Time creates a new Codec for the time.Time type.
// The encoded order is UTC time first, timezone offset second.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside of go's control (as they must be),
// and Time.Zone() can return names that will fail with Location.LoadLocation(name).
func Time() Codec[time.Time] { return internal.TimeCodec }

// BigInt creates a new Codec for the *big.Int type, with nils ordered first.
// This Codec may require a terminator when used within an aggregate Codec.
func BigInt() Codec[*big.Int] { return internal.BigIntCodec(true) }

// BigIntNilsLast creates a new Codec for the *big.Int type, with nils ordered last.
// This Codec may require a terminator when used within an aggregate Codec.
func BigIntNilsLast() Codec[*big.Int] { return internal.BigIntCodec(false) }

// BigFloat creates a new Codec for the *big.Float type, with nils ordered first.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// This Codec may require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the Accuracy.
func BigFloat() Codec[*big.Float] { return internal.BigFloatCodec(true) }

// BigFloatNilsLast creates a new Codec for the *big.Float type, with nils ordered last.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// This Codec may require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the Accuracy.
func BigFloatNilsLast() Codec[*big.Float] { return internal.BigFloatCodec(false) }

// BigRat creates a new Codec for the *big.Rat type, with nils ordered first.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec may require a terminator when used within an aggregate Codec.
func BigRat() Codec[*big.Rat] { return internal.BigRatCodec(true) }

// BigRatNilsLast creates a new Codec for the *big.Rat type, with nils ordered last.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec may require a terminator when used within an aggregate Codec.
func BigRatNilsLast() Codec[*big.Rat] { return internal.BigRatCodec(false) }

// Bytes creates a new Codec for []byte types, with nil slices ordered first.
// The encoded order is lexicographical.
// This Codec is more efficient than Codecs produced by SliceOf[[]byte],
// and will allow nil unlike String[string].
// This Codec requires a terminator when used within an aggregate Codec.
func Bytes[S ~[]byte]() Codec[S] {
	return internal.BytesCodec[S](true)
}

// BytesNilsLast creates a new Codec for []byte types, with nil slices ordered last.
// The encoded order is lexicographical.
// This Codec is more efficient than Codecs produced by SliceOfNilsLast[[]byte],
// and will allow nil unlike String[string].
// This Codec requires a terminator when used within an aggregate Codec.
func BytesNilsLast[S ~[]byte]() Codec[S] {
	return internal.BytesCodec[S](false)
}

// Codecs that delegate to other Codecs.

// PointerTo creates a new Codec for pointers to the type handled by elemCodec,
// with nils ordered first.
// Then encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec may require a terminator when used within an aggregate Codec.
func PointerTo[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerCodec[P](elemCodec, true)
}

// PointerToNilsLast creates a new Codec for pointers to the type handled by elemCodec,
// with nils ordered last.
// Then encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec may require a terminator when used within an aggregate Codec.
func PointerToNilsLast[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerCodec[P](elemCodec, false)
}

// ArrayOf creates a new Codec for the array type A with element type E.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// Arrays of different sizes are different types in go, and will require different Codecs.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// ArrayOf will panic if A is not an array type with element type E.
//
// This Codec makes heavy use of reflection, and should be avoided if possible.
func ArrayOf[A any, E any](elemCodec Codec[E]) Codec[A] {
	return internal.ArrayCodec[A](elemCodec)
}

// PointerToArrayOf creates a new Codec for pointers to the array type A with element type E,
// with nils ordered first.
// This Codec can be more efficient than Codecs produced by ArrayOf, depending on the size of the array.
// This Codec does not require a terminator when used within an aggregate Codec.
// Other than encoding a pointer value, this Codec behaves exactly like ArrayOf for non-nil values.
func PointerToArrayOf[P ~*A, A any, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerToArrayCodec[P](elemCodec, true)
}

// PointerToArrayOfNilsLast creates a new Codec for pointers to the array type A with element type E,
// with nils ordered last.
// This Codec can be more efficient than Codecs produced by ArrayOf, depending on the size of the array.
// This Codec does not require a terminator when used within an aggregate Codec.
// Other than encoding a pointer value, this Codec behaves exactly like ArrayOf for non-nil values.
func PointerToArrayOfNilsLast[P ~*A, A any, E any](elemCodec Codec[E]) Codec[P] {
	return internal.PointerToArrayCodec[P](elemCodec, false)
}

// SliceOf creates a new Codec for the slice type S with element type E, with nil slices ordered first.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOf[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	return internal.SliceCodec[S](elemCodec, true)
}

// SliceOfNilsLast creates a new Codec for the slice type S with element type E, with nil slices ordered last.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOfNilsLast[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	return internal.SliceCodec[S](elemCodec, false)
}

// MapOf creates a new Codec for the map type M using keyCodec and valueCodec, with nil maps ordered first.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return internal.MapCodec[M](keyCodec, valueCodec, true)
}

// MapOfNilsLast creates a new Codec for the map type M using keyCodec and valueCodec, with nil maps ordered last.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOfNilsLast[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	return internal.MapCodec[M](keyCodec, valueCodec, false)
}

// Negate returns a new Codec reversing the encoded order produced by codec.
// This Codec does not require a terminator when used within an aggregate Codec.
func Negate[T any](codec Codec[T]) Codec[T] {
	return internal.NegateCodec(codec)
}

// Codecs and functions for implementing new Codecs.

// Terminate returns a new Codec that escapes and terminates the encodings produced by codec.
// The returned Codec is NOT safe for concurrent access, and MUST be created anew when used.
func Terminate[T any](codec Codec[T]) Codec[T] {
	return internal.Terminate(codec)
}

// Terminate returns a new Codec that escapes and terminates the encodings produced by codec,
// if codec.RequiresTerminator() is true. Otherwise it returns codec.
// The returned Codec is NOT safe for concurrent access, and MUST be created anew when used.
func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	return internal.TerminateIfNeeded(codec)
}

// ReadPrefix is used to read the initial nil/empty/non-empty prefix byte by Codecs
// that encode types that can be nil or empty.
//
// nilable should be true if and only if nil is an allowed value of type T.
//
// emptyValue should point to the empty value of type T if it differs from the zero value of T.
//
// Returns done == false only if the value itself still needs to be read (neither nil nor empty),
// and there was no error reading the prefix. The returned value should be ignored in this case.
// Returns done == true otherwise, with the returned value being either nil or the empty value.
//
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
//
// See the PointerToStruct example for an example usage.
func ReadPrefix[T any](r io.Reader, nilable bool, emptyValue *T) (value T, done bool, err error) {
	return internal.ReadPrefix(r, nilable, emptyValue)
}

// WritePrefixNilsFirst is used to write the initial nil/empty/non-empty prefix byte by Codecs
// encoding types which can have nil or empty values, with nils ordered first.
//
// isNil, if non-nil, is a function returning whether a specific value of type T is nil.
// The functions IsNilPointer, IsNilSlice, and IsNilMap are provided for this purpose.
//
// isEmpty, if non-nil, is a function returning whether a specific value of type T is empty.
// The functions IsEmptyString, IsEmptySlice, and IsEmptyMap are provided for this purpose.
//
// Returns done == false only if the value itself still needs to be written (neither nil nor empty),
// and there was no error writing the prefix.
// Returns done == true otherwise (the prefix for nil or empty was written).
//
// Examples of types with differing nil and empty possibilities:
//
//	type     nil?  empty?
//	----------------------
//	int8     No    No
//	string   No    Yes
//	pointer  Yes   No
//	slice    Yes   Yes
//
// See the PointerToStruct example for an example usage.
func WritePrefixNilsFirst[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	return internal.WritePrefixNilsFirst(w, isNil, isEmpty, value)
}

// WritePrefixNilsLast is used to write the initial nil/empty/non-empty prefix byte by Codecs
// encoding types which can have nil or empty values, with nils ordered last.
// Otherwise it behaves exactly like WritePrefixNilsFirst.
func WritePrefixNilsLast[T any](w io.Writer, isNil, isEmpty func(T) bool, value T) (done bool, err error) {
	return internal.WritePrefixNilsLast(w, isNil, isEmpty, value)
}

// IsNilPointer is a helper function passed as the isNil argument in WritePrefixNilsFirst/Last.
func IsNilPointer[P ~*E, E any](value P) bool {
	return value == nil
}

// IsNilSlice is a helper function passed as the isNil argument in WritePrefixNilsFirst/Last.
func IsNilSlice[S ~[]E, E any](value S) bool {
	return value == nil
}

// IsNilMap is a helper function passed as the isNil argument in WritePrefixNilsFirst/Last.
func IsNilMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value == nil
}

// IsEmptyString is a helper function passed as the isEmpty argument in WritePrefixNilsFirst/Last.
func IsEmptyString[T ~string](value T) bool {
	return len(value) == 0
}

// IsEmptySlice is a helper function passed as the isEmpty argument in WritePrefixNilsFirst/Last.
func IsEmptySlice[S ~[]E, E any](value S) bool {
	return value != nil && len(value) == 0
}

// IsEmptyMap is a helper function passed as the isEmpty argument in WritePrefixNilsFirst/Last.
func IsEmptyMap[M ~map[K]V, K comparable, V any](value M) bool {
	return value != nil && len(value) == 0
}

// Convenience functions.

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
