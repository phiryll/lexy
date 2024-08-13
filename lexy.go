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

Functions returning Codecs for types that allow nil values return a [NillableCodec].
The Codecs returned by these functions will always order nil before all non-nil values.
Invoking [NillableCodec.NilsLast] on a NillableCodec will return a new Codec with same ordering,
except nils will be ordered after all non-nil values.

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

These Codec-returning functions require specifying a type parameter when invoked.
  - [Empty]
  - [MakeBool]
  - [MakeUint], [MakeUint8], [MakeUint16], [MakeUint32], [MakeUint64]
  - [MakeInt], [MakeInt8], [MakeInt16], [MakeInt32], [MakeInt64]
  - [MakeFloat32], [MakeFloat64]
  - [MakeString]
  - [MakeBytes]
  - [MakePointerTo], [MakeSliceOf], [MakeMapOf]

These are convenience functions using a []byte instead of an [io.Reader] or [io.Writer].
  - [Encode], [Decode]

These functions are used when creating custom Codecs.
  - [UnexpectedIfEOF]
  - [ReadPrefix], [WritePrefix]
*/
package lexy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"
)

// Codec defines a binary encoding for values of type T.
// Most of the Codec implementations provided by this package preserve the type's natural ordering,
// but nothing requires that behavior.
// The Read and Write methods should be lossless inverse operations.
// Exceptions to either of these behaviors should be clearly documented.
//
// All Codecs provided by lexy are safe for concurrent use if their delegate Codecs (if any) are.
type Codec[T any] interface {
	// Read reads from r and decodes a value of type T.
	//
	// Read will read from r until either it has all the data it needs, or EOF is reached.
	// If the returned error is non-nil, including [io.EOF], the returned value should be discarded.
	// Read will only return io.EOF if r returned io.EOF and no bytes were read.
	// Read will return [io.ErrUnexpectedEOF] if r returned io.EOF and a complete value was not successfully read.
	// Implementations of Read should never knowingly return an incomplete value.
	//
	// If instances of type T can be nil,
	// implementations of Read should invoke [ReadPrefix] as the first step,
	// and Write should invoke [WritePrefix].
	//
	// [io.Reader.Read] is permitted to return only immediately available data instead of waiting for more.
	// This may cause an error, or it may silently return incomplete data, depending on this Codec's implementation.
	// Implementations can use functions such as [io.Copy] and [io.ReadFull] to help avoid this problem.
	//
	// Read may repeatedly read small amounts of data from r,
	// so using a buffered io.Reader is recommended if appropriate.
	// Implementations of Read should never wrap r in a buffered io.Reader,
	// because doing so could consume excess data from r and corrupt following reads.
	Read(r io.Reader) (T, error)

	// Write encodes value and writes the encoded bytes to w.
	//
	// If instances of type T can be nil,
	// implementations of Write should invoke [WritePrefix] as the first step,
	// and Read should invoke [ReadPrefix].
	//
	// Write may repeatedly write small amounts of data to w,
	// so using a buffered io.Writer is recommended if appropriate.
	// Implementions of Write should not wrap w in a buffered io.Writer,
	// but if they do, the buffered io.Writer must be flushed before returning from Write.
	Write(w io.Writer, value T) error

	// RequiresTerminator returns whether data written by this Codec requires a terminator and escaping
	// when more data may be written to the same stream following the data written by this Codec.
	// This is true if either
	//   - Read may not know when to stop reading the data written by Write (strings, maps, some pointers, ...), or
	//   - Write could encode zero bytes for some value (strings, [Empty], ...).
	//
	// Users of this Codec must wrap it with [Terminate] or [TerminateIfNeeded] if RequiresTerminator may return true
	// and more data could be written following the data written by this Codec.
	// This is optional because terminating and escaping is unnecessary if this Codec should read until EOF,
	// and only the caller knows this.
	//
	// The Codec returned by [PointerTo] is unusual in that it only requires a terminator
	// if its referent Codec requires one.
	RequiresTerminator() bool
}

// A NillableCodec is a Codec where the value of type T can be nil.
// This interface exists to support the NilsLast method.
//
// In Go versions prior to 1.21, the compiler will not infer that a NillableCodec[T] is a Codec[T].
// However, an explicit cast works as expected, like this:
//
//	lexy.Terminate(lexy.Codec[[]string](lexy.SliceOf(lexy.String())))
//
// If Go cannot be upgraded to 1.21, a function like this might be helpful.
//
//	func castToCodec[T any](codec lexy.NillableCodec[T]) lexy.Codec[T] { return codec }
type NillableCodec[T any] interface {
	Codec[T]

	// NilsLast returns a Codec exactly like this Codec, but with nil values ordered last.
	NilsLast() NillableCodec[T]
}

// Codec instances for the common use cases.
// There are corresponding exported functions for each of these.
var (
	stdBool       Codec[bool]               = uintCodec[bool]{}
	stdUint       Codec[uint]               = asUint64Codec[uint]{}
	stdUint8      Codec[uint8]              = uintCodec[uint8]{}
	stdUint16     Codec[uint16]             = uintCodec[uint16]{}
	stdUint32     Codec[uint32]             = uintCodec[uint32]{}
	stdUint64     Codec[uint64]             = uintCodec[uint64]{}
	stdInt        Codec[int]                = asInt64Codec[int]{}
	stdInt8       Codec[int8]               = intCodec[int8]{math.MinInt8}
	stdInt16      Codec[int16]              = intCodec[int16]{math.MinInt16}
	stdInt32      Codec[int32]              = intCodec[int32]{math.MinInt32}
	stdInt64      Codec[int64]              = intCodec[int64]{math.MinInt64}
	stdFloat32    Codec[float32]            = float32Codec{}
	stdFloat64    Codec[float64]            = float64Codec{}
	stdComplex64  Codec[complex64]          = complex64Codec{}
	stdComplex128 Codec[complex128]         = complex128Codec{}
	stdString     Codec[string]             = stringCodec{}
	stdDuration   Codec[time.Duration]      = intCodec[time.Duration]{math.MinInt64}
	stdTime       Codec[time.Time]          = timeCodec{}
	stdBigInt     NillableCodec[*big.Int]   = bigIntCodec{true}
	stdBigFloat   NillableCodec[*big.Float] = bigFloatCodec{true}
	stdBigRat     NillableCodec[*big.Rat]   = bigRatCodec{true}
	stdBytes      NillableCodec[[]byte]     = bytesCodec[[]byte]{true}

	stdTermString   Codec[string]     = terminatorCodec[string]{stdString}
	stdTermBigFloat Codec[*big.Float] = terminatorCodec[*big.Float]{stdBigFloat}
	stdTermBytes    Codec[[]byte]     = terminatorCodec[[]byte]{stdBytes}
)

// Empty returns a Codec that reads and writes no data.
// [Codec.Read] returns the zero value of T.
// Codec.Read and [Codec.Write] will never return an error, including [io.EOF].
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

// BigInt returns a NillableCodec for the *big.Int type, with nils ordered first.
// This Codec does not require a terminator when used within an aggregate Codec.
func BigInt() NillableCodec[*big.Int] { return stdBigInt }

// BigFloat returns a NillableCodec for the *big.Float type, with nils ordered first.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// Like floats, -Inf, -0.0, +0.0, and +Inf all have a big.Float representation.
// However, there is no big.Float representation for NaN.
// This Codec requires a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the value's [big.Accuracy].
func BigFloat() NillableCodec[*big.Float] { return stdBigFloat }

// TerminatedBigFloat returns a Codec for the *big.Float type which escapes and terminates the encoded bytes.
// This Codec does not require a terminator when used within an aggregate Codec.
func TerminatedBigFloat() Codec[*big.Float] { return stdTermBigFloat }

// BigRat returns a NillableCodec for the *big.Rat type, with nils ordered first.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec does not require a terminator when used within an aggregate Codec.
func BigRat() NillableCodec[*big.Rat] { return stdBigRat }

// Bytes returns a NillableCodec for the []byte type, with nil slices ordered first.
// A []byte is written as-is following a nil/non-nil indicator.
// This Codec is more efficient than Codecs produced by [SliceOf]([Uint8]()),
// and will allow nil unlike [String].
// This Codec requires a terminator when used within an aggregate Codec.
func Bytes() NillableCodec[[]byte] { return stdBytes }

// TerminatedBytes returns a Codec for the []byte type which escapes and terminates the encoded bytes.
// This Codec does not require a terminator when used within an aggregate Codec.
func TerminatedBytes() Codec[[]byte] { return stdTermBytes }

// PointerTo returns a NillableCodec for the *E type, with nil pointers ordered first.
// The encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec requires a terminator when used within an aggregate Codec if elemCodec does.
func PointerTo[E any](elemCodec Codec[E]) NillableCodec[*E] {
	return MakePointerTo[*E](elemCodec)
}

// SliceOf returns a NillableCodec for the []E type, with nil slices ordered first.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOf[E any](elemCodec Codec[E]) NillableCodec[[]E] {
	return MakeSliceOf[[]E](elemCodec)
}

// MapOf returns a NillableCodec for the map[K]V type, with nil maps ordered first.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOf[K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) NillableCodec[map[K]V] {
	return MakeMapOf[map[K]V](keyCodec, valueCodec)
}

// Negate returns a Codec reversing the encoded order of codec.
// This Codec does not require a terminator when used within an aggregate Codec.
func Negate[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	// Negate must escape and terminate its delegate whether it requires it or not,
	// but shouldn't wrap if the delegate is already a terminatorCodec.
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
	if codec == nil {
		panic("codec must be non-nil")
	}
	return terminatorCodec[T]{codec}
}

// TerminateIfNeeded returns a Codec that escapes and terminates the encodings produced by codec,
// if [Codec.RequiresTerminator] returns true for codec. Otherwise it returns codec.
func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	// This also covers the case if codec is a terminator.
	if !codec.RequiresTerminator() {
		return codec
	}
	return terminatorCodec[T]{codec}
}

// Functions to help in implementing new Codecs.

// UnexpectedIfEOF returns [io.ErrUnexpectedEOF] if err is [io.EOF], and returns err otherwise.
//
// This helps make [Codec.Read] implementations easier to read.
// See the examples for usage patterns.
func UnexpectedIfEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Prefixes to use for encodings for types whose instances can be nil.
// The values were chosen so that prefixNilFirst < prefixNonNil < prefixNilLast,
// and neither the prefixes nor their complements need to be escaped.
const (
	// Room for more between prefixNonNil and prefixNilLast if needed.
	prefixNilFirst byte = 0x02
	prefixNonNil   byte = 0x03
	prefixNilLast  byte = 0xFD
)

// Convenience byte slices.
var (
	pNilFirst = []byte{prefixNilFirst}
	pNonNil   = []byte{prefixNonNil}
	pNilLast  = []byte{prefixNilLast}
)

// ReadPrefix is used to read the initial nil/non-nil prefix byte from r by Codecs
// that encode types whose instances can be nil.
// Invoking ReadPrefix should the first action taken by [Codec.Read] for these Codecs,
// since it allows an early return if the decoded value is nil.
// This is a typical usage:
//
//	func (c someCodecType) Read(r io.Reader) (T, error) {
//	    if done, err := lexy.ReadPrefix(r); done {
//	        return nil, err
//	    }
//	    // read, decode, and return a non-nil value
//	}
//
// ReadPrefix returns done == false only if a non-nil value still needs to be read and decoded,
// and there was no error reading the prefix.
//
// If ReadPrefix returns done == true, then the caller is done reading this value
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was read.
//
// ReadPrefix will return [io.EOF] only if no bytes were read and [io.Reader.Read] returned io.EOF.
// ReadPrefix will not return an error if a prefix was successfully read and io.Reader.Read returned io.EOF,
// because the read of the prefix was successful.
// Any subsequent read from r should properly return 0 bytes read and io.EOF.
//
//nolint:nonamedreturns
func ReadPrefix(r io.Reader) (done bool, err error) {
	prefix := []byte{0}
	_, err = io.ReadFull(r, prefix)
	if err != nil {
		return true, err
	}
	switch prefix[0] {
	case prefixNilFirst, prefixNilLast:
		return true, nil
	case prefixNonNil:
		return false, nil
	default:
		//nolint:err113
		return true, fmt.Errorf("unexpected prefix %X", prefix[0])
	}
}

// WritePrefix is used to write the initial nil/non-nil prefix byte to w by Codecs
// that encode types whose instances can be nil.
// The prefix written depends on the values of isNil and nilsFirst.
// Invoking WritePrefix should be the first action taken by [Codec.Write] for these Codecs,
// since it allows an early return if value is nil.
// This is a typical usage:
//
//	func (c someCodecType) Write(w io.Writer, value T) error {
//	    if done, err := lexy.WritePrefix(w, value == nil, true); done {
//	        return err
//	    }
//	    // encode and write the non-nil value
//	}
//
// WritePrefix returns done == false only if isNil is false and there was no error writing the prefix,
// in which case the caller still needs to write the non-nil value to w.
//
// If WritePrefix returns done == true, then the caller is done writing the value to w
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was successfully written.
//
//nolint:nonamedreturns
func WritePrefix(w io.Writer, isNil, nilsFirst bool) (done bool, err error) {
	var prefix []byte
	switch {
	case !isNil:
		prefix = pNonNil
	case nilsFirst:
		prefix = pNilFirst
	default:
		prefix = pNilLast
	}
	if _, err := w.Write(prefix); err != nil {
		return true, err
	}
	return isNil, nil
}

// Convenience functions.

// The default size when allocating a buffer, chosen because it should fit in a cache line.
const defaultBufSize = 64

// Encode returns value encoded using codec as a []byte.
//
// This is a convenience function.
// Use [Codec.Write] when encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufSize))
	if err := codec.Write(buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode returns a value decoded from data using codec.
//
// This is a convenience function.
// Use [Codec.Read] when decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return codec.Read(bytes.NewReader(data))
}
