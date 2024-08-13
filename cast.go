package lexy

import (
	"io"
	"math"
)

// Codecs for types with different underlying types.
// These merely delegate to the Codecs for the underlying types and cast.
// Previous version of lexy had generic definitions for the Codecs with the logic.
// While that does make the execution path slightly faster for non-underlying types,
// it also creates a copy of the entire implementation for every type.
// The casting wrapper types here should take up a lot less space.

// MakeBool returns a Codec for a type with an underlying type of bool.
// Other than the underlying type, this is the same as [Bool].
func MakeBool[T ~bool]() Codec[T] { return uintCodec[T]{} }

// MakeUint returns a Codec for a type with an underlying type of uint.
// Other than the underlying type, this is the same as [Uint].
func MakeUint[T ~uint]() Codec[T] { return asUint64Codec[T]{} }

// MakeUint8 returns a Codec for a type with an underlying type of uint8.
// Other than the underlying type, this is the same as [Uint8].
func MakeUint8[T ~uint8]() Codec[T] { return uintCodec[T]{} }

// MakeUint16 returns a Codec for a type with an underlying type of uint16.
// Other than the underlying type, this is the same as [Uint16].
func MakeUint16[T ~uint16]() Codec[T] { return uintCodec[T]{} }

// MakeUint32 returns a Codec for a type with an underlying type of uint32.
// Other than the underlying type, this is the same as [Uint32].
func MakeUint32[T ~uint32]() Codec[T] { return uintCodec[T]{} }

// MakeUint64 returns a Codec for a type with an underlying type of uint64.
// Other than the underlying type, this is the same as [Uint64].
func MakeUint64[T ~uint64]() Codec[T] { return uintCodec[T]{} }

// MakeInt returns a Codec for a type with an underlying type of int.
// Other than the underlying type, this is the same as [Int].
func MakeInt[T ~int]() Codec[T] { return asInt64Codec[T]{} }

// MakeInt8 returns a Codec for a type with an underlying type of int8.
// Other than the underlying type, this is the same as [Int8].
func MakeInt8[T ~int8]() Codec[T] { return intCodec[T]{math.MinInt8} }

// MakeInt16 returns a Codec for a type with an underlying type of int16.
// Other than the underlying type, this is the same as [Int16].
func MakeInt16[T ~int16]() Codec[T] { return intCodec[T]{math.MinInt16} }

// MakeInt32 returns a Codec for a type with an underlying type of int32.
// Other than the underlying type, this is the same as [Int32].
func MakeInt32[T ~int32]() Codec[T] { return intCodec[T]{math.MinInt32} }

// MakeInt64 returns a Codec for a type with an underlying type of int64.
// Other than the underlying type, this is the same as [Int64].
func MakeInt64[T ~int64]() Codec[T] { return intCodec[T]{math.MinInt64} }

// MakeFloat32 returns a Codec for a type with an underlying type of float32.
// Other than the underlying type, this is the same as [Float32].
func MakeFloat32[T ~float32]() Codec[T] { return castFloat32Codec[T]{} }

// MakeFloat64 returns a Codec for a type with an underlying type of float64.
// Other than the underlying type, this is the same as [Float64].
func MakeFloat64[T ~float64]() Codec[T] { return castFloat64Codec[T]{} }

// MakeString returns a Codec for a type with an underlying type of string.
// Other than the underlying type, this is the same as [String].
func MakeString[T ~string]() Codec[T] { return stringCodec[T]{} }

// MakeBytes returns a NillableCodec for a type with an underlying type of []byte, with nil slices ordered first.
// Other than the underlying type, this is the same as [Bytes].
func MakeBytes[S ~[]byte]() NillableCodec[S] { return bytesCodec[S]{true} }

// MakePointerTo returns a NillableCodec for a type with an underlying type of *E, with nil pointers ordered first.
// Other than the underlying type, this is the same as [PointerTo].
func MakePointerTo[P ~*E, E any](elemCodec Codec[E]) NillableCodec[P] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return pointerCodec[P, E]{elemCodec, true}
}

// MakeSliceOf returns a NillableCodec for a type with an underlying type of []E, with nil slices ordered first.
// Other than the underlying type, this is the same as [SliceOf].
func MakeSliceOf[S ~[]E, E any](elemCodec Codec[E]) NillableCodec[S] {
	if elemCodec == nil {
		panic("elemCodec must be non-nil")
	}
	return sliceCodec[S, E]{TerminateIfNeeded(elemCodec), true}
}

// MakeMapOf returns a NillableCodec for a type with an underlying type of map[K]V, with nil maps ordered first.
// Other than the underlying type, this is the same as [MapOf].
func MakeMapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) NillableCodec[M] {
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

// It would be really nice to have just one castCodec[T ~U, U any],
// but that's not possible in Go.

type castFloat32Codec[T ~float32] struct{}

func (castFloat32Codec[T]) Read(r io.Reader) (T, error) {
	value, err := stdFloat32Codec.Read(r)
	return T(value), err
}

func (castFloat32Codec[T]) Write(w io.Writer, value T) error {
	return stdFloat32Codec.Write(w, float32(value))
}

func (castFloat32Codec[T]) RequiresTerminator() bool {
	return stdFloat32Codec.RequiresTerminator()
}

type castFloat64Codec[T ~float64] struct{}

func (castFloat64Codec[T]) Read(r io.Reader) (T, error) {
	value, err := stdFloat64Codec.Read(r)
	return T(value), err
}

func (castFloat64Codec[T]) Write(w io.Writer, value T) error {
	return stdFloat64Codec.Write(w, float64(value))
}

func (castFloat64Codec[T]) RequiresTerminator() bool {
	return stdFloat64Codec.RequiresTerminator()
}
