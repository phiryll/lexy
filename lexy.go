/*
Package lexy defines an API for lexicographically ordered binary encodings.
Implementations are provided for most builtin Go data types,
and supporting functions are provided to allow the creation of custom encodings.

The [Codec][T] interface defines an encoding, with methods to encode and decode values of type T.
Functions returning Codecs for different types constitute the majority of this API.
There are two kinds of Codec-returning functions defined by this package,
those for which Go can infer the type arguments, and those for which Go cannot.
The former have terser names, as in [Int16]().
The latter have names starting with "Make", as in [MakeInt16][MyIntType]().
These latter functions are only needed when creating a Codec for a type that is not the same as its underlying type.
[Empty] also requires a type argument when used and is the only exception to this naming convention.

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
  - [BigInt], [BigFloat], [BigRat], [TerminatedBigFloat]
  - [Bytes], [TerminatedBytes]
  - [PointerTo], [SliceOf], [MapOf]
  - [Negate]
  - [Terminate], [TerminateIfNeeded]
  - [NilsLast]

These Codec-returning functions require specifying a type parameter when invoked.
  - [Empty]
  - [MakeBool]
  - [MakeUint], [MakeUint8], [MakeUint16], [MakeUint32], [MakeUint64]
  - [MakeInt], [MakeInt8], [MakeInt16], [MakeInt32], [MakeInt64]
  - [MakeFloat32], [MakeFloat64]
  - [MakeString]
  - [MakeBytes]
  - [MakePointerTo], [MakeSliceOf], [MakeMapOf]
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

	// Put encodes value into buf, returning the number of bytes written.
	//
	// Put will panic if buf is too small, and still may have written some data to buf.
	// Put will write only the bytes that encode value.
	Put(buf []byte, value T) int

	// Get decodes a value of type T from buf, returning the value and the number of bytes read.
	//
	// If buf is empty and this Codec could encode zero bytes for some value,
	// Get will return that value and 0 bytes read.
	// If buf is empty and this Codec cannot encode zero bytes for any value,
	// Get will return the zero value of T and a byte count < 0.
	// Checking the returned byte count is the only way to distinguish these cases.
	// Get will panic if a value of type T cannot be successfully decoded from a non-empty buf.
	// Get will not modify buf.
	Get(buf []byte) (T, int)

	// RequiresTerminator returns whether encoded values require a terminator and escaping
	// if more data is written following the encoded value.
	// This is the case for unbounded types like strings and slices,
	// as well as types whose encodings can be zero bytes.
	// Types whose encodings are always a fixed size, like integers and floats,
	// never require a terminator and escaping.
	//
	// Users of this Codec must wrap it with [Terminate] or [TerminateIfNeeded] if RequiresTerminator may return true
	// and more data could be written following the data written by this Codec.
	// This is optional because terminating and escaping is unnecessary
	// if the use of this Codec should decode entire buffer.
	//
	// The Codec returned by [PointerTo] is unusual in that it only requires a terminator
	// if its referent Codec requires one.
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
// This Codec requires a terminator when used within an aggregate Codec.
func Empty[T any]() Codec[T] { return emptyCodec[T]{} }

// Bool returns a Codec for the bool type.
// The encoded order is false, then true.
// This Codec does not require a terminator when used within an aggregate Codec.
func Bool() Codec[bool] { return stdBool }

// Uint returns a Codec for the uint type.
// Values are converted to/from uint64 and encoded with [Uint64].
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint() Codec[uint] { return stdUint }

// Uint8 returns a Codec for the uint8 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint8() Codec[uint8] { return stdUint8 }

// Uint16 returns a Codec for the uint16 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint16() Codec[uint16] { return stdUint16 }

// Uint32 returns a Codec for the uint32 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint32() Codec[uint32] { return stdUint32 }

// Uint64 returns a Codec for the uint64 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint64() Codec[uint64] { return stdUint64 }

// Int returns a Codec for the int type.
// Values are converted to/from int64 and encoded with [Int64].
// This Codec does not require a terminator when used within an aggregate Codec.
func Int() Codec[int] { return stdInt }

// Int8 returns a Codec for the int8 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int8() Codec[int8] { return stdInt8 }

// Int16 returns a Codec for the int16 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int16() Codec[int16] { return stdInt16 }

// Int32 returns a Codec for the int32 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int32() Codec[int32] { return stdInt32 }

// Int64 returns a Codec for the int64 type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int64() Codec[int64] { return stdInt64 }

// Float32 returns a Codec for the float32 type.
// All bits of the value are preserved by this encoding.
// There are many different bit patterns for NaN, and their encodings will be distinct.
// No ordering distinction is made between quiet and signaling NaNs.
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
func Float32() Codec[float32] { return stdFloat32 }

// Float64 returns a Codec for the float64 type.
// Other than handling float64 instances, this function behaves the same as [Float32].
func Float64() Codec[float64] { return stdFloat64 }

// Complex64 returns a Codec for the complex64 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for [Float32].
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex64() Codec[complex64] { return stdComplex64 }

// Complex128 returns a Codec for the complex128 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for [Float64].
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex128() Codec[complex128] { return stdComplex128 }

// String returns a Codec for the string type.
// This Codec requires a terminator when used within an aggregate Codec.
//
// A string is encoded as its bytes.
// This encoded order may be surprising.
// A string in Go is essentially an immutable []byte without any text semantics.
// For a UTF-8 string, the order is the same as the lexicographical order of the Unicode code points.
// However, even this is not intuitive. For example, 'Z' < 'a'.
// Collation is locale-dependent, and any ordering could be incorrect in another locale.
func String() Codec[string] { return stdString }

// TerminatedString returns a Codec for the string type which escapes and terminates the encoded bytes.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// This is a convenience function, it returns the same Codec as [Terminate]([String]()).
func TerminatedString() Codec[string] { return stdTermString }

// Duration returns a Codec for the time.Duration type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Duration() Codec[time.Duration] { return stdDuration }

// Time returns a Codec for the time.Time type.
// The encoded order is UTC time first, timezone offset second.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside of Go's control (as they must be),
// and [time.Time.Zone] can return names that will fail with [time.LoadLocation] in the same program.
func Time() Codec[time.Time] { return stdTime }

// BigInt returns a Codec for the *big.Int type, with nils ordered first.
// This Codec requires a terminator when used within an aggregate Codec.
func BigInt() Codec[*big.Int] { return stdBigInt }

// BigFloat returns a Codec for the *big.Float type, with nils ordered first.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// Like floats, -Inf, -0.0, +0.0, and +Inf all have a big.Float representation.
// However, there is no big.Float representation for NaN.
// This Codec requires a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the value's [big.Accuracy].
func BigFloat() Codec[*big.Float] { return stdBigFloat }

// BigRat returns a Codec for the *big.Rat type, with nils ordered first.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec does not require a terminator when used within an aggregate Codec.
func BigRat() Codec[*big.Rat] { return stdBigRat }

// Bytes returns a Codec for the []byte type, with nil slices ordered first.
// A []byte is written as-is following a nil/non-nil indicator.
// This Codec is more efficient than Codecs produced by [SliceOf]([Uint8]()),
// and will allow nil unlike [String].
// This Codec requires a terminator when used within an aggregate Codec.
func Bytes() Codec[[]byte] { return stdBytes }

// TerminatedBytes returns a Codec for the []byte type which escapes and terminates the encoded bytes.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// This is a convenience function, it returns the same Codec as [Terminate]([Bytes]()).
func TerminatedBytes() Codec[[]byte] { return stdTermBytes }

// PointerTo returns a Codec for the *E type, with nil pointers ordered first.
// The encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec requires a terminator when used within an aggregate Codec if elemCodec does.
func PointerTo[E any](elemCodec Codec[E]) Codec[*E] {
	elemCodec.RequiresTerminator() // force panic if nil
	return pointerCodec[E]{elemCodec, PrefixNilsFirst}
}

// SliceOf returns a Codec for the []E type, with nil slices ordered first.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOf[E any](elemCodec Codec[E]) Codec[[]E] {
	return sliceCodec[E]{TerminateIfNeeded(elemCodec), PrefixNilsFirst}
}

// MapOf returns a Codec for the map[K]V type, with nil maps ordered first.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOf[K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[map[K]V] {
	return mapCodec[K, V]{
		TerminateIfNeeded(keyCodec),
		TerminateIfNeeded(valueCodec),
		PrefixNilsFirst,
	}
}

// Negate returns a Codec reversing the encoded order of codec.
// This Codec does not require a terminator when used within an aggregate Codec.
func Negate[T any](codec Codec[T]) Codec[T] {
	// Negate must escape and terminate its delegate whether it requires it or not,
	// but shouldn't wrap if the delegate is already a terminatorCodec.
	// This will also attempt to wrap a nil Codec, causing Terminate() to panic.
	if _, ok := codec.(terminatorCodec[T]); !ok {
		codec = Terminate(codec)
	}
	return negateCodec[T]{codec}
}

// Terminate returns a Codec that escapes and terminates the encodings produced by codec.
// This function is for the rare edge case requiring a Codec's encodings to be escaped and terminated,
// whether or not it normally requires it.
// Most of the time, [TerminateIfNeeded] should be used instead.
func Terminate[T any](codec Codec[T]) Codec[T] {
	codec.RequiresTerminator() // force panic if nil
	return terminatorCodec[T]{codec}
}

// TerminateIfNeeded returns a Codec that escapes and terminates the encodings produced by codec,
// if [Codec.RequiresTerminator] returns true for codec. Otherwise it returns codec.
func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	// This also covers the case if codec is a terminator.
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
// Codecs returned [Terminate], [TerminateIfNeeded], and [Negate] will cause NilsLast to panic,
// regardless of the Codec they are wrapping.
func NilsLast[T any](codec Codec[T]) Codec[T] {
	if c, ok := codec.(nillableCodec[T]); ok {
		return c.nilsLast()
	}
	panic(badTypeError{codec})
}

// Functions to help in implementing new Codecs.

// TODO: Add more after looking for pain points in examples.

// Helper functions used by implementations.

// The default size when allocating a buffer, chosen because it should fit in a cache line.
const defaultBufSize = 64

// mustNonNil panics with a nilError with the given name if x is nil.
// The best way to panic if something is nil is to use it,
// use this function only if that isn't possible.
func mustNonNil(x any, name string) {
	if x == nil {
		panic(nilError{name})
	}
}

// mustCopy is like the built-in copy(dst, src),
// except that it panics if dst is not large enough to hold all of src.
// mustCopy returns the number of bytes copied, which is len(src).
func mustCopy(dst, src []byte) int {
	if len(src) == 0 {
		return 0
	}
	_ = dst[len(src)-1]
	return copy(dst, src)
}
