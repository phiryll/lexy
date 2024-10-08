/*
Package lexy defines an API for lexicographically ordered binary encodings.
Implementations are provided for most builtin Go data types,
and supporting functionality is provided to allow for the creation of user-defined encodings.

The [Codec][T] interface defines an encoding, with methods to encode and decode values of type T.
Functions returning Codecs for different types constitute the majority of this API.
There are two kinds of Codec-returning functions defined by this package,
those for which Go can infer the type arguments, and those for which Go cannot.
The former have terser names, as in [Int16].
The latter have names starting with "Cast", as in [CastInt16][MyIntType].
These latter functions are only needed when creating a Codec for a type that is not the same as its underlying type.
[Empty] also requires a type argument when used and is the only exception to this naming convention.

All Codecs provided by lexy are safe for concurrent use if their delegate Codecs (if any) are.

All Codecs provided by lexy will order nils first if nil can be encoded.
Invoking [NilsLast](codec) on a Codec will return a Codec which orders nils last,
but only for the pointer, slice, map, []byte, and *big.Int/Float/Rat Codecs provided by lexy.

See [Codec.RequiresTerminator] for details on when escaping and terminating encoded bytes is required.

These Codec-returning functions do not require specifying a type parameter when invoked.
  - [Bool]
  - [Uint], [Uint8], [Uint16], [Uint32], [Uint64]
  - [Int], [Int8], [Int16], [Int32], [Int64]
  - [Float32], [Float64]
  - [Complex64], [Complex128]
  - [String], [TerminatedString]
  - [Time], [Duration]
  - [BigInt], [BigFloat], [BigRat]
  - [Bytes], [TerminatedBytes]
  - [PointerTo], [SliceOf], [MapOf]
  - [Negate]
  - [Terminate]
  - [NilsLast]

These Codec-returning functions require specifying a type parameter when invoked.
  - [Empty]
  - [CastBool]
  - [CastUint], [CastUint8], [CastUint16], [CastUint32], [CastUint64]
  - [CastInt], [CastInt8], [CastInt16], [CastInt32], [CastInt64]
  - [CastFloat32], [CastFloat64]
  - [CastString]
  - [CastBytes]
  - [CastPointerTo], [CastSliceOf], [CastMapOf]

These are implementations of [Prefix], used when creating user-defined Codecs
that can encode types whose instances can be nil.
  - [PrefixNilsFirst], [PrefixNilsLast]
*/
package lexy

import (
	"math/big"
	"time"
)

// Codec defines a binary encoding for values of type T.
// Most of the Codec implementations provided by this package preserve the type's natural ordering,
// but nothing requires that behavior.
// Append and Put should produce the same encoded bytes.
// Get must be able to decode encodings produced by Append and Put.
// Encoding and decoding should be lossless inverse operations.
// Exceptions to any of these behaviors are allowed, but should be clearly documented.
//
// All Codecs provided by lexy will order nils first if instances of type T can be nil.
// Invoking [NilsLast](codec) on a Codec will return a Codec which orders nils last,
// but only for the pointer, slice, map, []byte, and *big.Int/Float/Rat Codecs provided by lexy.
//
// If instances of type T can be nil,
// implementations should invoke the appropriate method of [PrefixNilsFirst] or [PrefixNilsLast]
// as the first step of encoding or decoding method implementations.
// See the [Prefix] docs for example usage idioms.
//
// All Codecs provided by lexy are safe for concurrent use if their delegate Codecs (if any) are.
type Codec[T any] interface {
	// Append encodes value and appends the encoded bytes to buf, returning the updated buffer.
	//
	// If buf is nil and no bytes are appended, Append may return nil.
	Append(buf []byte, value T) []byte

	// Put encodes value into buf, returning buf following what was written.
	//
	// Put will panic if buf is too small, and still may have written some data to buf.
	// Put will write only the bytes that encode value.
	Put(buf []byte, value T) []byte

	// Get decodes a value of type T from buf, returning the value and buf following the encoded value.
	// Get will panic if a value of type T cannot be successfully decoded from buf.
	// If buf is empty and this Codec could encode zero bytes for some value, Get will return that value and buf.
	// Get will not modify buf.
	Get(buf []byte) (T, []byte)

	// RequiresTerminator returns whether encoded values require escaping and a terminator
	// if more data is written following the encoded value.
	// This is the case for most unbounded types like slices and maps,
	// as well as types whose encodings can be zero bytes.
	// Wrapping this Codec with [Terminate] will return a Codec which behaves properly in these situations.
	//
	// For the rest of this doc comment, "requires escaping" is shorthand for
	// "requires escaping and a terminator if more data is written following the encoded value."
	//
	// Codecs that could encode zero bytes, like those for string and [Empty], always require escaping.
	// Codecs that could produce two distinct non-empty encodings with one being a prefix of the other,
	// like those for slices and maps, always require escaping.
	// Codecs that cannot produce two distinct non-empty encodings with one being a prefix of the other,
	// like those for primitive integers and floats, never require escaping.
	// Codecs that always encode to a non-zero fixed number of bytes are a special case of this.
	//
	// The net effect of escaping and terminating is to prevent one encoding from being the prefix of another,
	// while maintaining the same lexicographical ordering.
	RequiresTerminator() bool
}

// Codec instances for the common use cases.
// There are corresponding exported functions for each of these.
var (
	stdBool       Codec[bool]          = boolCodec{}
	stdUint       Codec[uint]          = castUint64[uint]{}
	stdUint8      Codec[uint8]         = uint8Codec{}
	stdUint16     Codec[uint16]        = uint16Codec{}
	stdUint32     Codec[uint32]        = uint32Codec{}
	stdUint64     Codec[uint64]        = uint64Codec{}
	stdInt        Codec[int]           = castInt64[int]{}
	stdInt8       Codec[int8]          = int8Codec{}
	stdInt16      Codec[int16]         = int16Codec{}
	stdInt32      Codec[int32]         = int32Codec{}
	stdInt64      Codec[int64]         = int64Codec{}
	stdFloat32    Codec[float32]       = float32Codec{}
	stdFloat64    Codec[float64]       = float64Codec{}
	stdComplex64  Codec[complex64]     = complex64Codec{}
	stdComplex128 Codec[complex128]    = complex128Codec{}
	stdString     Codec[string]        = stringCodec{}
	stdDuration   Codec[time.Duration] = castInt64[time.Duration]{}
	stdTime       Codec[time.Time]     = timeCodec{}
	stdBigFloat   Codec[*big.Float]    = bigFloatCodec{PrefixNilsFirst}
	stdBigInt     Codec[*big.Int]      = bigIntCodec{PrefixNilsFirst}
	stdBigRat     Codec[*big.Rat]      = bigRatCodec{PrefixNilsFirst}
	stdBytes      Codec[[]byte]        = bytesCodec{PrefixNilsFirst}

	stdTermString Codec[string] = terminatorCodec[string]{stdString}
	stdTermBytes  Codec[[]byte] = terminatorCodec[[]byte]{stdBytes}
)

// Empty returns a Codec that encodes instances of T to zero bytes.
// Get returns the zero value of T.
// No method of this Codec will ever fail.
//
// This is useful for empty structs, which are often used as map values.
// This Codec requires escaping, as defined by [Codec.RequiresTerminator].
func Empty[T any]() Codec[T] { return emptyCodec[T]{} }

// Bool returns a Codec for the bool type.
// The encoded order is false, then true.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Bool() Codec[bool] { return stdBool }

// Uint returns a Codec for the uint type.
// Values are converted to/from uint64 and encoded with [Uint64].
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Uint() Codec[uint] { return stdUint }

// Uint8 returns a Codec for the uint8 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Uint8() Codec[uint8] { return stdUint8 }

// Uint16 returns a Codec for the uint16 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Uint16() Codec[uint16] { return stdUint16 }

// Uint32 returns a Codec for the uint32 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Uint32() Codec[uint32] { return stdUint32 }

// Uint64 returns a Codec for the uint64 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Uint64() Codec[uint64] { return stdUint64 }

// Int returns a Codec for the int type.
// Values are converted to/from int64 and encoded with [Int64].
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Int() Codec[int] { return stdInt }

// Int8 returns a Codec for the int8 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Int8() Codec[int8] { return stdInt8 }

// Int16 returns a Codec for the int16 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Int16() Codec[int16] { return stdInt16 }

// Int32 returns a Codec for the int32 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Int32() Codec[int32] { return stdInt32 }

// Int64 returns a Codec for the int64 type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Int64() Codec[int64] { return stdInt64 }

// Float32 returns a Codec for the float32 type.
// All bits of the value are preserved by this encoding.
// There are many different bit patterns for NaN, and their encodings will be distinct.
// No ordering distinction is made between quiet and signaling NaNs.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
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
func Float32() Codec[float32] { return stdFloat32 }

// Float64 returns a Codec for the float64 type.
// Other than handling float64 instances, this function behaves the same as [Float32].
func Float64() Codec[float64] { return stdFloat64 }

// Complex64 returns a Codec for the complex64 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for [Float32].
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Complex64() Codec[complex64] { return stdComplex64 }

// Complex128 returns a Codec for the complex128 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for [Float64].
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Complex128() Codec[complex128] { return stdComplex128 }

// String returns a Codec for the string type.
// This Codec requires escaping, as defined by [Codec.RequiresTerminator].
//
// A string is encoded as its bytes.
// This encoded order may be surprising.
// A string in Go is essentially an immutable []byte without any text semantics.
// For a UTF-8 string, the order is the same as the lexicographical order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent, and any ordering could be incorrect in another locale.
func String() Codec[string] { return stdString }

// TerminatedString returns a Codec for the string type which escapes and terminates the encoded bytes.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
//
// This is a convenience function, it returns the same Codec as [Terminate]([String]()).
func TerminatedString() Codec[string] { return stdTermString }

// Time returns a Codec for the time.Time type.
// The encoded order is UTC time first, timezone offset second.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
//
// This Codec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside of Go's control (as they must be),
// and [time.Time.Zone] can return names that will fail with [time.LoadLocation] in the same program.
func Time() Codec[time.Time] { return stdTime }

// Duration returns a Codec for the time.Duration type.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Duration() Codec[time.Duration] { return stdDuration }

// BigInt returns a Codec for the *big.Int type, with nils ordered first.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func BigInt() Codec[*big.Int] { return stdBigInt }

// BigFloat returns a Codec for the *big.Float type, with nils ordered first.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// Like floats, -Inf, -0.0, +0.0, and +Inf all have a big.Float representation.
// However, there is no big.Float representation for NaN.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
//
// This Codec is lossy. It does not encode the value's [big.Accuracy].
func BigFloat() Codec[*big.Float] { return stdBigFloat }

// BigRat returns a Codec for the *big.Rat type, with nils ordered first.
// The encoded order is signed numerator first, positive denominator second.
// Note that this is not the natural ordering for rational numbers.
// big.Rat will normalize its value to lowest terms.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func BigRat() Codec[*big.Rat] { return stdBigRat }

// Bytes returns a Codec for the []byte type, with nil slices ordered first.
// A []byte is written as-is following a nil/non-nil indicator.
// This Codec is more efficient than Codecs produced by [SliceOf]([Uint8]()),
// and will allow nil unlike [String].
// This Codec requires escaping, as defined by [Codec.RequiresTerminator].
func Bytes() Codec[[]byte] { return stdBytes }

// TerminatedBytes returns a Codec for the []byte type which escapes and terminates the encoded bytes.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
//
// This is a convenience function, it returns the same Codec as [Terminate]([Bytes]()).
func TerminatedBytes() Codec[[]byte] { return stdTermBytes }

// PointerTo returns a Codec for the *E type, with nil pointers ordered first.
// The encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec requires escaping if elemCodec does, as defined by [Codec.RequiresTerminator].
func PointerTo[E any](elemCodec Codec[E]) Codec[*E] {
	elemCodec.RequiresTerminator() // force panic if nil
	return pointerCodec[E]{elemCodec, PrefixNilsFirst}
}

// SliceOf returns a Codec for the []E type, with nil slices ordered first.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires escaping, as defined by [Codec.RequiresTerminator].
func SliceOf[E any](elemCodec Codec[E]) Codec[[]E] {
	return sliceCodec[E]{Terminate(elemCodec), PrefixNilsFirst}
}

// MapOf returns a Codec for the map[K]V type, with nil maps ordered first.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires escaping, as defined by [Codec.RequiresTerminator].
func MapOf[K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[map[K]V] {
	return mapCodec[K, V]{
		Terminate(keyCodec),
		Terminate(valueCodec),
		PrefixNilsFirst,
	}
}

// Negate returns a Codec reversing the encoded order of codec.
// This Codec does not require escaping, as defined by [Codec.RequiresTerminator].
func Negate[T any](codec Codec[T]) Codec[T] {
	// negateEscapeCodec internally escapes its data, so unwrap any terminatorCodecs.
	for {
		delegate, ok := codec.(terminatorCodec[T])
		if !ok {
			break
		}
		codec = delegate.codec
	}
	if codec.RequiresTerminator() {
		return negateEscapeCodec[T]{codec}
	}
	return negateCodec[T]{codec}
}

// Terminate returns a Codec that escapes and terminates the encodings produced by codec,
// if [Codec.RequiresTerminator] returns true for codec. Otherwise it returns codec.
func Terminate[T any](codec Codec[T]) Codec[T] {
	// This also covers the case if codec is a terminatorCodec.
	if !codec.RequiresTerminator() {
		return codec
	}
	return terminatorCodec[T]{codec}
}

// Unexported interface with an unexported method for NilsLast to use.
// This can only be implemented by Codecs in this package.
// This is by far the cleanest way to implement NilsLast(codec),
// type switching with generics is just not supported very well.
type nillableCodec[T any] interface {
	nilsLast() Codec[T]
}

// NilsLast returns a Codec exactly like codec, but with nils ordered last.
// NilsLast will panic if codec is not a pointer, slice, map, []byte, or *big.Int/Float/Rat Codec provided by lexy.
// Codecs returned by [Negate] and [Terminate] will cause NilsLast to panic,
// regardless of the Codec they are wrapping.
func NilsLast[T any](codec Codec[T]) Codec[T] {
	if c, ok := codec.(nillableCodec[T]); ok {
		return c.nilsLast()
	}
	panic(badTypeError{codec})
}

// Helper functionality used by implementations.

// copyAll is like the built-in copy(dst, src),
// except that it panics if dst is not large enough to hold all of src.
// copyAll returns a slice into dst following what was written.
func copyAll(dst, src []byte) []byte {
	if len(src) == 0 {
		return dst
	}
	_ = dst[len(src)-1]
	return dst[copy(dst, src):]
}

// extend ensures that n bytes can be appended to buf without another allocation,
// returning the resulting slice. This was copied from slices.Grow (added in go 1.21).
func extend(buf []byte, n int) []byte {
	if n -= cap(buf) - len(buf); n > 0 {
		buf = append(buf[:cap(buf)], make([]byte, n)...)[:len(buf)]
	}
	return buf
}

const bitsPerByte = 8

func numBytes(numBits int) int {
	return (numBits + bitsPerByte - 1) / bitsPerByte
}
