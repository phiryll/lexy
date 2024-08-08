// Package lexy defines an API for lexicographically ordered binary encodings.
// Implementations are provided for most builtin go data types,
// and supporting functions are provided to allow clients to create custom encodings.
package lexy

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"
)

// A Codec defines a lexicographically ordered binary encoding for values of a data type.
//
// Encoded values should normally have the same order as the values they encode.
// The Read and Write methods should be lossless inverse operations.
// Exceptions to either of these behaviors should be clearly documented.
//
// All Codecs provided by lexy are safe for concurrent use if their delegate Codecs (if any) are.
type Codec[T any] interface {
	// Read reads from r and decodes a value of type T.
	//
	// Read will read from r until either it has all the data it needs, or r stops returning data.
	// r.Read is permitted to return only immediately available data instead of waiting for more.
	// This may cause an error, or it may silently return incomplete data, depending on this Codec's implementation.
	// Implementations of Read should never knowingly return incomplete data.
	//
	// If the returned error is non-nil, including io.EOF, the returned value should be discarded.
	// Read will only return io.EOF if r returned io.EOF and no bytes were read.
	// Read will return io.ErrUnexpectedEOF if r returned io.EOF and a complete value was not successfully read.
	//
	// If instances of type T can be nil,
	// implementations of Read should invoke [ReadPrefix] as the first step,
	// and Write should invoke [WritePrefix].
	//
	// Read may repeatedly read small amounts of data from r,
	// so using a buffered io.Reader is recommended if appropriate.
	// Implementations of Read should never wrap r in a buffered io.Reader,
	// because doing so could consume excess data from r.
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

	// RequiresTerminator returns whether encoded data written by this Codec requires a terminator,
	// and therefore also must be escaped, if more data is written following the encoded data.
	// Stated another way, RequiresTerminator must return true if Read may not know
	// when to stop reading the data encoded by Write,
	// or if Write could encode zero bytes for some value.
	// This is the case for unbounded types like strings, slices, and maps,
	// as well as empty struct types.
	//
	// Users of this Codec must wrap it with Terminate or TerminateIfNeeded if RequiresTerminator may return true
	// and more data could be written following the encoded data.
	// For example, lexy's slice Codec implementation must wrap its element Codec with TerminateIfNeeded.
	// A user does not need to consider wrapping this Codec if either:
	//	- this Codec is known to not require it, and will never require it (the int16 Codec, e.g.), or
	//	- the data written by this Codec will always be at the end of the stream when read.
	//
	// Lexy's pointer Codec implementation is an unusual use case in that it only requires a terminator
	// if its element Codec requires one.
	// This is only because the pointer Codec encodes at most a single element,
	// and does not itself encode any data following that element.
	RequiresTerminator() bool
}

// Codecs used by other Codecs.
var (
	stdUint32Codec  Codec[uint32]  = uintCodec[uint32]{}
	stdUint64Codec  Codec[uint64]  = uintCodec[uint64]{}
	stdInt8Codec    Codec[int8]    = intCodec[int8]{signBit: math.MinInt8}
	stdInt32Codec   Codec[int32]   = intCodec[int32]{signBit: math.MinInt32}
	stdInt64Codec   Codec[int64]   = intCodec[int64]{signBit: math.MinInt64}
	stdFloat32Codec Codec[float32] = float32Codec[float32]{}
	stdFloat64Codec Codec[float64] = float64Codec[float64]{}
)

// Codecs that do not delegate to other Codecs, for types with builtin underlying types.

// Empty creates a new Codec that reads and writes no data.
// Read returns the zero value of T.
// Read and Write will never return an error, including io.EOF.
// This is useful for empty structs, which are often used as map values.
// This Codec requires a terminator when used within an aggregate Codec.
func Empty[T any]() Codec[T] { return emptyCodec[T]{} }

// Bool creates a new Codec for a type with an underlying type of bool.
// This Codec does not require a terminator when used within an aggregate Codec.
func Bool[T ~bool]() Codec[T] { return uintCodec[T]{} }

// Uint creates a new Codec for a type with an underlying type of uint.
// Values are converted to/from uint64 and encoded with Uint64[uint64]().
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint[T ~uint]() Codec[T] { return asUint64Codec[T]{} }

// Uint8 creates a new Codec for a type with an underlying type of uint8.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint8[T ~uint8]() Codec[T] { return uintCodec[T]{} }

// Uint16 creates a new Codec for a type with an underlying type of uint16.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint16[T ~uint16]() Codec[T] { return uintCodec[T]{} }

// Uint32 creates a new Codec for a type with an underlying type of uint32.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint32[T ~uint32]() Codec[T] { return uintCodec[T]{} }

// Uint64 creates a new Codec for a type with an underlying type of uint64.
// This Codec does not require a terminator when used within an aggregate Codec.
func Uint64[T ~uint64]() Codec[T] { return uintCodec[T]{} }

// Int creates a new Codec for a type with an underlying type of int.
// Values are converted to/from int64 and encoded with Int64[int64]().
// This Codec does not require a terminator when used within an aggregate Codec.
func Int[T ~int]() Codec[T] { return asInt64Codec[T]{} }

// Int8 creates a new Codec for a type with an underlying type of int8.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int8[T ~int8]() Codec[T] { return intCodec[T]{signBit: math.MinInt8} }

// Int16 creates a new Codec for a type with an underlying type of int16.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int16[T ~int16]() Codec[T] { return intCodec[T]{signBit: math.MinInt16} }

// Int32 creates a new Codec for a type with an underlying type of int32.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int32[T ~int32]() Codec[T] { return intCodec[T]{signBit: math.MinInt32} }

// Int64 creates a new Codec for a type with an underlying type of int64.
// This Codec does not require a terminator when used within an aggregate Codec.
func Int64[T ~int64]() Codec[T] { return intCodec[T]{signBit: math.MinInt64} }

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
func Float32[T ~float32]() Codec[T] { return float32Codec[T]{} }

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
func Float64[T ~float64]() Codec[T] { return float64Codec[T]{} }

// Complex64 returns the Codec for the complex64 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for Float32.
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex64() Codec[complex64] { return complex64Codec{} }

// Complex128 returns the Codec for the complex128 type.
// The encoded order is real part first, imaginary part second,
// with those parts ordered as documented for Float64.
// This Codec does not require a terminator when used within an aggregate Codec.
func Complex128() Codec[complex128] { return complex128Codec{} }

// String creates a new Codec for a type with an underlying type of string.
// This Codec requires a terminator when used within an aggregate Codec.
func String[T ~string]() Codec[T] { return stringCodec[T]{} }

// Duration creates a new Codec for the time.Duration type.
// This Codec does not require a terminator when used within an aggregate Codec.
func Duration() Codec[time.Duration] { return Int64[time.Duration]() }

// Codecs that do not delegate to other Codecs, for types without builtin underlying types (all structs).

// Time creates a new Codec for the time.Time type.
// The encoded order is UTC time first, timezone offset second.
// This Codec does not require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It encodes the timezone's offset, but not its name.
// It will therefore lose information about Daylight Saving Time.
// Timezone names and DST behavior are defined outside of go's control (as they must be),
// and Time.Zone() can return names that will fail with Location.LoadLocation(name).
func Time() Codec[time.Time] { return timeCodec{} }

// BigInt creates a new Codec for the *big.Int type, with nils ordered first.
// This Codec may require a terminator when used within an aggregate Codec.
func BigInt() Codec[*big.Int] { return bigIntCodec{true} }

// BigIntNilsLast creates a new Codec for the *big.Int type, with nils ordered last.
// This Codec may require a terminator when used within an aggregate Codec.
func BigIntNilsLast() Codec[*big.Int] { return bigIntCodec{false} }

// BigFloat creates a new Codec for the *big.Float type, with nils ordered first.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// This Codec may require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the Accuracy.
func BigFloat() Codec[*big.Float] { return bigFloatCodec{true} }

// BigFloatNilsLast creates a new Codec for the *big.Float type, with nils ordered last.
// The encoded order is the numeric value first, precision second, and rounding mode third.
// This Codec may require a terminator when used within an aggregate Codec.
//
// This Codec is lossy. It does not encode the Accuracy.
func BigFloatNilsLast() Codec[*big.Float] { return bigFloatCodec{false} }

// BigRat creates a new Codec for the *big.Rat type, with nils ordered first.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec may require a terminator when used within an aggregate Codec.
func BigRat() Codec[*big.Rat] { return bigRatCodec{true} }

// BigRatNilsLast creates a new Codec for the *big.Rat type, with nils ordered last.
// The encoded order is signed numerator first, positive denominator second.
// Note that big.Rat will normalize its value to lowest terms.
// This Codec may require a terminator when used within an aggregate Codec.
func BigRatNilsLast() Codec[*big.Rat] { return bigRatCodec{false} }

// Bytes creates a new Codec for []byte types, with nil slices ordered first.
// The encoded order is lexicographical.
// This Codec is more efficient than Codecs produced by SliceOf[[]byte],
// and will allow nil unlike String[string].
// This Codec requires a terminator when used within an aggregate Codec.
func Bytes[S ~[]byte]() Codec[S] { return bytesCodec[S]{true} }

// BytesNilsLast creates a new Codec for []byte types, with nil slices ordered last.
// The encoded order is lexicographical.
// This Codec is more efficient than Codecs produced by SliceOfNilsLast[[]byte],
// and will allow nil unlike String[string].
// This Codec requires a terminator when used within an aggregate Codec.
func BytesNilsLast[S ~[]byte]() Codec[S] { return bytesCodec[S]{false} }

// Codecs that delegate to other Codecs.

// PointerTo creates a new Codec for pointers to the type handled by elemCodec,
// with nils ordered first.
// Then encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec may require a terminator when used within an aggregate Codec.
func PointerTo[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return pointerCodec[P, E]{elemCodec, true}
}

// PointerToNilsLast creates a new Codec for pointers to the type handled by elemCodec,
// with nils ordered last.
// Then encoded order of non-nil values is the same as is produced by elemCodec.
// This Codec may require a terminator when used within an aggregate Codec.
func PointerToNilsLast[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return pointerCodec[P, E]{elemCodec, false}
}

// SliceOf creates a new Codec for the slice type S with element type E, with nil slices ordered first.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOf[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, E]{TerminateIfNeeded(elemCodec), true}
}

// SliceOfNilsLast creates a new Codec for the slice type S with element type E, with nil slices ordered last.
// The encoded order is lexicographical using the encoded order of elemCodec for the elements.
// This Codec requires a terminator when used within an aggregate Codec.
func SliceOfNilsLast[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, E]{TerminateIfNeeded(elemCodec), false}
}

// MapOf creates a new Codec for the map type M using keyCodec and valueCodec, with nil maps ordered first.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return mapCodec[M, K, V]{
		TerminateIfNeeded(keyCodec),
		TerminateIfNeeded(valueCodec),
		true,
	}
}

// MapOfNilsLast creates a new Codec for the map type M using keyCodec and valueCodec, with nil maps ordered last.
// The encoded order for non-nil maps is empty maps first, with all other maps randomly ordered after.
// This Codec requires a terminator when used within an aggregate Codec.
func MapOfNilsLast[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return mapCodec[M, K, V]{
		TerminateIfNeeded(keyCodec),
		TerminateIfNeeded(valueCodec),
		false,
	}
}

// Negate returns a new Codec reversing the encoded order produced by codec.
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

// Codecs and functions to help in implementing new Codecs.

// UnexpectedIfEOF returns io.ErrUnexpectedEOF if err is io.EOF, and returns err otherwise.
//
// This helps make Codec.Read implementations a little easier to read.
func UnexpectedIfEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Terminate returns a new Codec that escapes and terminates the encodings produced by codec.
func Terminate[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	return terminatorCodec[T]{codec: codec}
}

// Terminate returns a new Codec that escapes and terminates the encodings produced by codec,
// if codec.RequiresTerminator() is true. Otherwise it returns codec.
func TerminateIfNeeded[T any](codec Codec[T]) Codec[T] {
	if codec == nil {
		panic("codec must be non-nil")
	}
	// This also covers the case if codec is a terminator.
	if !codec.RequiresTerminator() {
		return codec
	}
	return terminatorCodec[T]{codec: codec}
}

// Prefixes to use for encodings for types whose instances can be nil.
// The values were chosen so that nils-first < non-nil < nils-last,
// and neither the prefixes nor their complements need to be escaped.
const (
	// Room for more between non-nil and nils-last if needed.
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
// since it allows an early return if the value read is nil.
// This is a typical usage:
//
//	func (c someCodecType) Read(r io.Reader) (T, error) {
//	    if done, err := lexy.ReadPrefix(r); done {
//	        return nil, err
//	    }
//	    // read, decode, and return a non-nil value
//	}
//
// ReadPrefix returns done == false only if the non-nil value still needs to be read,
// and there was no error reading the prefix.
//
// If ReadPrefix returns done == true, then the caller is done reading this value
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was read.
//
// ReadPrefix will return io.EOF only if no bytes were read and r.Read returned io.EOF.
// ReadPrefix will not return an error if a prefix was successfully read and r.Read returned io.EOF,
// because the read of the prefix was successful.
// Any subsequent read from r by the caller will properly return 0 bytes read and io.EOF.
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
		return true, fmt.Errorf("unexpected prefix %X", prefix[0])
	}
}

// WritePrefix writes a nil/non-nil prefix byte to w based on the values of isNil and nilsFirst.
// Invoking WritePrefix should the first action taken by [Codec.Write] for these Codecs,
// since it allows an early return if the value written is nil.
// This is a typical usage:
//
//	func (c someCodecType) Write(w io.Writer, value T) error {
//	    if done, err := lexy.WritePrefix(w, value == nil, true); done {
//	        return err
//	    }
//	    // encode and write non-nil value
//	}
//
// WritePrefix returns done == false only if isNil is false and there was no error writing the prefix,
// in which case the caller still needs to write the non-nil value to w.
//
// If WritePrefix returns done == true, then the caller is done writing the current value to w
// regardless of the returned error value.
// Either there was an error, or there was no error and the nil prefix was successfully written.
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

// Encode returns value encoded using codec as a new []byte.
//
// This is a convenience function.
// Use Codec.Write when encoding multiple values to the same byte stream.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	if err := codec.Write(buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode returns a decoded value from a []byte using codec.
//
// This is a convenience function.
// Use Codec.Read when decoding multiple values from the same byte stream.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	return codec.Read(bytes.NewReader(data))
}
