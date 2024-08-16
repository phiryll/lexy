package lexy

import (
	"io"
)

// Codecs for types with different underlying types.
// These merely delegate to the Codecs for the underlying types and cast.
// Previous version of lexy had generic definitions for the Codecs with the logic.
// While that does make the execution path slightly faster for non-underlying types,
// it also creates a copy of the entire implementation for every type.
// The casting wrapper types here should take up a lot less space.

// MakeBool returns a Codec for a type with an underlying type of bool.
// Other than the underlying type, this is the same as [Bool].
func MakeBool[T ~bool]() Codec[T] { return castBool[T]{} }

// MakeUint returns a Codec for a type with an underlying type of uint.
// Other than the underlying type, this is the same as [Uint].
func MakeUint[T ~uint]() Codec[T] { return castUint64[T]{} }

// MakeUint8 returns a Codec for a type with an underlying type of uint8.
// Other than the underlying type, this is the same as [Uint8].
func MakeUint8[T ~uint8]() Codec[T] { return castUint8[T]{} }

// MakeUint16 returns a Codec for a type with an underlying type of uint16.
// Other than the underlying type, this is the same as [Uint16].
func MakeUint16[T ~uint16]() Codec[T] { return castUint16[T]{} }

// MakeUint32 returns a Codec for a type with an underlying type of uint32.
// Other than the underlying type, this is the same as [Uint32].
func MakeUint32[T ~uint32]() Codec[T] { return castUint32[T]{} }

// MakeUint64 returns a Codec for a type with an underlying type of uint64.
// Other than the underlying type, this is the same as [Uint64].
func MakeUint64[T ~uint64]() Codec[T] { return castUint64[T]{} }

// MakeInt returns a Codec for a type with an underlying type of int.
// Other than the underlying type, this is the same as [Int].
func MakeInt[T ~int]() Codec[T] { return castInt64[T]{} }

// MakeInt8 returns a Codec for a type with an underlying type of int8.
// Other than the underlying type, this is the same as [Int8].
func MakeInt8[T ~int8]() Codec[T] { return castInt8[T]{} }

// MakeInt16 returns a Codec for a type with an underlying type of int16.
// Other than the underlying type, this is the same as [Int16].
func MakeInt16[T ~int16]() Codec[T] { return castInt16[T]{} }

// MakeInt32 returns a Codec for a type with an underlying type of int32.
// Other than the underlying type, this is the same as [Int32].
func MakeInt32[T ~int32]() Codec[T] { return castInt32[T]{} }

// MakeInt64 returns a Codec for a type with an underlying type of int64.
// Other than the underlying type, this is the same as [Int64].
func MakeInt64[T ~int64]() Codec[T] { return castInt64[T]{} }

// MakeFloat32 returns a Codec for a type with an underlying type of float32.
// Other than the underlying type, this is the same as [Float32].
func MakeFloat32[T ~float32]() Codec[T] { return castFloat32[T]{} }

// MakeFloat64 returns a Codec for a type with an underlying type of float64.
// Other than the underlying type, this is the same as [Float64].
func MakeFloat64[T ~float64]() Codec[T] { return castFloat64[T]{} }

// MakeString returns a Codec for a type with an underlying type of string.
// Other than the underlying type, this is the same as [String].
func MakeString[T ~string]() Codec[T] { return castString[T]{} }

// MakeBytes returns a NillableCodec for a type with an underlying type of []byte, with nil slices ordered first.
// Other than the underlying type, this is the same as [Bytes].
func MakeBytes[S ~[]byte]() NillableCodec[S] { return castBytes[S]{stdBytes} }

// MakePointerTo returns a NillableCodec for a type with an underlying type of *E, with nil pointers ordered first.
// Other than the underlying type, this is the same as [PointerTo].
func MakePointerTo[P ~*E, E any](elemCodec Codec[E]) NillableCodec[P] {
	return castPointer[P, E]{PointerTo(elemCodec)}
}

// MakeSliceOf returns a NillableCodec for a type with an underlying type of []E, with nil slices ordered first.
// Other than the underlying type, this is the same as [SliceOf].
func MakeSliceOf[S ~[]E, E any](elemCodec Codec[E]) NillableCodec[S] {
	return castSlice[S, E]{SliceOf(elemCodec)}
}

// MakeMapOf returns a NillableCodec for a type with an underlying type of map[K]V, with nil maps ordered first.
// Other than the underlying type, this is the same as [MapOf].
func MakeMapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) NillableCodec[M] {
	return castMap[M, K, V]{MapOf(keyCodec, valueCodec)}
}

// It would be really nice to have just one castCodec[T ~U, U any],
// but that's not possible in Go.

type (
	castBool[T ~bool]             struct{}
	castUint8[T ~uint8]           struct{}
	castUint16[T ~uint16]         struct{}
	castUint32[T ~uint32]         struct{}
	castUint64[T ~uint64 | ~uint] struct{}
	castInt8[T ~int8]             struct{}
	castInt16[T ~int16]           struct{}
	castInt32[T ~int32]           struct{}
	castInt64[T ~int64 | ~int]    struct{}
	castFloat32[T ~float32]       struct{}
	castFloat64[T ~float64]       struct{}
	castString[T ~string]         struct{}
	castBytes[T ~[]byte]          struct {
		codec NillableCodec[[]byte]
	}
	castPointer[P ~*E, E any] struct {
		codec NillableCodec[*E]
	}
	castSlice[S ~[]E, E any] struct {
		codec NillableCodec[[]E]
	}
	castMap[M ~map[K]V, K comparable, V any] struct {
		codec NillableCodec[map[K]V]
	}
)

func (castBool[T]) Write(w io.Writer, value T) error {
	return stdBool.Write(w, bool(value))
}

func (castBool[T]) Read(r io.Reader) (T, error) {
	value, err := stdBool.Read(r)
	return T(value), err
}

func (castBool[T]) RequiresTerminator() bool {
	return stdBool.RequiresTerminator()
}

func (castUint8[T]) Write(w io.Writer, value T) error {
	return stdUint8.Write(w, uint8(value))
}

func (castUint8[T]) Read(r io.Reader) (T, error) {
	value, err := stdUint8.Read(r)
	return T(value), err
}

func (castUint8[T]) RequiresTerminator() bool {
	return stdUint8.RequiresTerminator()
}

func (castUint16[T]) Write(w io.Writer, value T) error {
	return stdUint16.Write(w, uint16(value))
}

func (castUint16[T]) Read(r io.Reader) (T, error) {
	value, err := stdUint16.Read(r)
	return T(value), err
}

func (castUint16[T]) RequiresTerminator() bool {
	return stdUint16.RequiresTerminator()
}

func (castUint32[T]) Write(w io.Writer, value T) error {
	return stdUint32.Write(w, uint32(value))
}

func (castUint32[T]) Read(r io.Reader) (T, error) {
	value, err := stdUint32.Read(r)
	return T(value), err
}

func (castUint32[T]) RequiresTerminator() bool {
	return stdUint32.RequiresTerminator()
}

func (castUint64[T]) Write(w io.Writer, value T) error {
	return stdUint64.Write(w, uint64(value))
}

func (castUint64[T]) Read(r io.Reader) (T, error) {
	value, err := stdUint64.Read(r)
	return T(value), err
}

func (castUint64[T]) RequiresTerminator() bool {
	return stdUint64.RequiresTerminator()
}

func (castInt8[T]) Write(w io.Writer, value T) error {
	return stdInt8.Write(w, int8(value))
}

func (castInt8[T]) Read(r io.Reader) (T, error) {
	value, err := stdInt8.Read(r)
	return T(value), err
}

func (castInt8[T]) RequiresTerminator() bool {
	return stdInt8.RequiresTerminator()
}

func (castInt16[T]) Write(w io.Writer, value T) error {
	return stdInt16.Write(w, int16(value))
}

func (castInt16[T]) Read(r io.Reader) (T, error) {
	value, err := stdInt16.Read(r)
	return T(value), err
}

func (castInt16[T]) RequiresTerminator() bool {
	return stdInt16.RequiresTerminator()
}

func (castInt32[T]) Write(w io.Writer, value T) error {
	return stdInt32.Write(w, int32(value))
}

func (castInt32[T]) Read(r io.Reader) (T, error) {
	value, err := stdInt32.Read(r)
	return T(value), err
}

func (castInt32[T]) RequiresTerminator() bool {
	return stdInt32.RequiresTerminator()
}

func (castInt64[T]) Write(w io.Writer, value T) error {
	return stdInt64.Write(w, int64(value))
}

func (castInt64[T]) Read(r io.Reader) (T, error) {
	value, err := stdInt64.Read(r)
	return T(value), err
}

func (castInt64[T]) RequiresTerminator() bool {
	return stdInt64.RequiresTerminator()
}

func (castFloat32[T]) Write(w io.Writer, value T) error {
	return stdFloat32.Write(w, float32(value))
}

func (castFloat32[T]) Read(r io.Reader) (T, error) {
	value, err := stdFloat32.Read(r)
	return T(value), err
}

func (castFloat32[T]) RequiresTerminator() bool {
	return stdFloat32.RequiresTerminator()
}

func (castFloat64[T]) Write(w io.Writer, value T) error {
	return stdFloat64.Write(w, float64(value))
}

func (castFloat64[T]) Read(r io.Reader) (T, error) {
	value, err := stdFloat64.Read(r)
	return T(value), err
}

func (castFloat64[T]) RequiresTerminator() bool {
	return stdFloat64.RequiresTerminator()
}

func (castString[T]) Write(w io.Writer, value T) error {
	return stdString.Write(w, string(value))
}

func (castString[T]) Read(r io.Reader) (T, error) {
	value, err := stdString.Read(r)
	return T(value), err
}

func (castString[T]) RequiresTerminator() bool {
	return stdString.RequiresTerminator()
}

func (c castBytes[T]) Write(w io.Writer, value T) error {
	return c.codec.Write(w, value)
}

func (c castBytes[T]) Read(r io.Reader) (T, error) {
	return c.codec.Read(r)
}

func (c castBytes[T]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

func (c castBytes[T]) NilsLast() NillableCodec[T] {
	return castBytes[T]{c.codec.NilsLast()}
}

func (c castPointer[P, E]) Write(w io.Writer, value P) error {
	return c.codec.Write(w, (*E)(value))
}

func (c castPointer[P, E]) Read(r io.Reader) (P, error) {
	return c.codec.Read(r)
}

func (c castPointer[P, E]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

func (c castPointer[P, E]) NilsLast() NillableCodec[P] {
	return castPointer[P, E]{c.codec.NilsLast()}
}

func (c castSlice[S, E]) Write(w io.Writer, value S) error {
	return c.codec.Write(w, []E(value))
}

func (c castSlice[S, E]) Read(r io.Reader) (S, error) {
	return c.codec.Read(r)
}

func (c castSlice[S, E]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

func (c castSlice[S, E]) NilsLast() NillableCodec[S] {
	return castSlice[S, E]{c.codec.NilsLast()}
}

func (c castMap[M, K, V]) Write(w io.Writer, value M) error {
	return c.codec.Write(w, map[K]V(value))
}

func (c castMap[M, K, V]) Read(r io.Reader) (M, error) {
	return c.codec.Read(r)
}

func (c castMap[M, K, V]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

func (c castMap[M, K, V]) NilsLast() NillableCodec[M] {
	return castMap[M, K, V]{c.codec.NilsLast()}
}
