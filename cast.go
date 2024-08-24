package lexy

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

// MakeBytes returns a Codec for a type with an underlying type of []byte, with nil slices ordered first.
// Other than the underlying type, this is the same as [Bytes].
func MakeBytes[S ~[]byte]() Codec[S] {
	//nolint:forcetypeassert
	return castBytes[S]{stdBytes.(bytesCodec)}
}

// MakePointerTo returns a Codec for a type with an underlying type of *E, with nil pointers ordered first.
// Other than the underlying type, this is the same as [PointerTo].
func MakePointerTo[P ~*E, E any](elemCodec Codec[E]) Codec[P] {
	//nolint:forcetypeassert
	return castPointer[P, E]{PointerTo(elemCodec).(pointerCodec[E])}
}

// MakeSliceOf returns a Codec for a type with an underlying type of []E, with nil slices ordered first.
// Other than the underlying type, this is the same as [SliceOf].
func MakeSliceOf[S ~[]E, E any](elemCodec Codec[E]) Codec[S] {
	//nolint:forcetypeassert
	return castSlice[S, E]{SliceOf(elemCodec).(sliceCodec[E])}
}

// MakeMapOf returns a Codec for a type with an underlying type of map[K]V, with nil maps ordered first.
// Other than the underlying type, this is the same as [MapOf].
func MakeMapOf[M ~map[K]V, K comparable, V any](keyCodec Codec[K], valueCodec Codec[V]) Codec[M] {
	//nolint:forcetypeassert
	return castMap[M, K, V]{MapOf(keyCodec, valueCodec).(mapCodec[K, V])}
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
		codec bytesCodec
	}
	castPointer[P ~*E, E any] struct {
		codec pointerCodec[E]
	}
	castSlice[S ~[]E, E any] struct {
		codec sliceCodec[E]
	}
	castMap[M ~map[K]V, K comparable, V any] struct {
		codec mapCodec[K, V]
	}
)

func (castBool[T]) Append(buf []byte, value T) []byte {
	return stdBool.Append(buf, bool(value))
}

func (castBool[T]) Put(buf []byte, value T) int {
	return stdBool.Put(buf, bool(value))
}

func (castBool[T]) Get(buf []byte) (T, int) {
	value, n := stdBool.Get(buf)
	return T(value), n
}

func (castBool[T]) RequiresTerminator() bool {
	return stdBool.RequiresTerminator()
}

func (castUint8[T]) Append(buf []byte, value T) []byte {
	return stdUint8.Append(buf, uint8(value))
}

func (castUint8[T]) Put(buf []byte, value T) int {
	return stdUint8.Put(buf, uint8(value))
}

func (castUint8[T]) Get(buf []byte) (T, int) {
	value, n := stdUint8.Get(buf)
	return T(value), n
}

func (castUint8[T]) RequiresTerminator() bool {
	return stdUint8.RequiresTerminator()
}

func (castUint16[T]) Append(buf []byte, value T) []byte {
	return stdUint16.Append(buf, uint16(value))
}

func (castUint16[T]) Put(buf []byte, value T) int {
	return stdUint16.Put(buf, uint16(value))
}

func (castUint16[T]) Get(buf []byte) (T, int) {
	value, n := stdUint16.Get(buf)
	return T(value), n
}

func (castUint16[T]) RequiresTerminator() bool {
	return stdUint16.RequiresTerminator()
}

func (castUint32[T]) Append(buf []byte, value T) []byte {
	return stdUint32.Append(buf, uint32(value))
}

func (castUint32[T]) Put(buf []byte, value T) int {
	return stdUint32.Put(buf, uint32(value))
}

func (castUint32[T]) Get(buf []byte) (T, int) {
	value, n := stdUint32.Get(buf)
	return T(value), n
}

func (castUint32[T]) RequiresTerminator() bool {
	return stdUint32.RequiresTerminator()
}

func (castUint64[T]) Append(buf []byte, value T) []byte {
	return stdUint64.Append(buf, uint64(value))
}

func (castUint64[T]) Put(buf []byte, value T) int {
	return stdUint64.Put(buf, uint64(value))
}

func (castUint64[T]) Get(buf []byte) (T, int) {
	value, n := stdUint64.Get(buf)
	return T(value), n
}

func (castUint64[T]) RequiresTerminator() bool {
	return stdUint64.RequiresTerminator()
}

func (castInt8[T]) Append(buf []byte, value T) []byte {
	return stdInt8.Append(buf, int8(value))
}

func (castInt8[T]) Put(buf []byte, value T) int {
	return stdInt8.Put(buf, int8(value))
}

func (castInt8[T]) Get(buf []byte) (T, int) {
	value, n := stdInt8.Get(buf)
	return T(value), n
}

func (castInt8[T]) RequiresTerminator() bool {
	return stdInt8.RequiresTerminator()
}

func (castInt16[T]) Append(buf []byte, value T) []byte {
	return stdInt16.Append(buf, int16(value))
}

func (castInt16[T]) Put(buf []byte, value T) int {
	return stdInt16.Put(buf, int16(value))
}

func (castInt16[T]) Get(buf []byte) (T, int) {
	value, n := stdInt16.Get(buf)
	return T(value), n
}

func (castInt16[T]) RequiresTerminator() bool {
	return stdInt16.RequiresTerminator()
}

func (castInt32[T]) Append(buf []byte, value T) []byte {
	return stdInt32.Append(buf, int32(value))
}

func (castInt32[T]) Put(buf []byte, value T) int {
	return stdInt32.Put(buf, int32(value))
}

func (castInt32[T]) Get(buf []byte) (T, int) {
	value, n := stdInt32.Get(buf)
	return T(value), n
}

func (castInt32[T]) RequiresTerminator() bool {
	return stdInt32.RequiresTerminator()
}

func (castInt64[T]) Append(buf []byte, value T) []byte {
	return stdInt64.Append(buf, int64(value))
}

func (castInt64[T]) Put(buf []byte, value T) int {
	return stdInt64.Put(buf, int64(value))
}

func (castInt64[T]) Get(buf []byte) (T, int) {
	value, n := stdInt64.Get(buf)
	return T(value), n
}

func (castInt64[T]) RequiresTerminator() bool {
	return stdInt64.RequiresTerminator()
}

func (castFloat32[T]) Append(buf []byte, value T) []byte {
	return stdFloat32.Append(buf, float32(value))
}

func (castFloat32[T]) Put(buf []byte, value T) int {
	return stdFloat32.Put(buf, float32(value))
}

func (castFloat32[T]) Get(buf []byte) (T, int) {
	value, n := stdFloat32.Get(buf)
	return T(value), n
}

func (castFloat32[T]) RequiresTerminator() bool {
	return stdFloat32.RequiresTerminator()
}

func (castFloat64[T]) Append(buf []byte, value T) []byte {
	return stdFloat64.Append(buf, float64(value))
}

func (castFloat64[T]) Put(buf []byte, value T) int {
	return stdFloat64.Put(buf, float64(value))
}

func (castFloat64[T]) Get(buf []byte) (T, int) {
	value, n := stdFloat64.Get(buf)
	return T(value), n
}

func (castFloat64[T]) RequiresTerminator() bool {
	return stdFloat64.RequiresTerminator()
}

func (castString[T]) Append(buf []byte, value T) []byte {
	return stdString.Append(buf, string(value))
}

func (castString[T]) Put(buf []byte, value T) int {
	return stdString.Put(buf, string(value))
}

func (castString[T]) Get(buf []byte) (T, int) {
	value, n := stdString.Get(buf)
	return T(value), n
}

func (castString[T]) RequiresTerminator() bool {
	return stdString.RequiresTerminator()
}

func (c castBytes[T]) Append(buf []byte, value T) []byte {
	return c.codec.Append(buf, []byte(value))
}

func (c castBytes[T]) Put(buf []byte, value T) int {
	return c.codec.Put(buf, []byte(value))
}

func (c castBytes[T]) Get(buf []byte) (T, int) {
	return c.codec.Get(buf)
}

func (c castBytes[T]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c castBytes[T]) nilsLast() Codec[T] {
	//nolint:forcetypeassert
	return castBytes[T]{c.codec.nilsLast().(bytesCodec)}
}

func (c castPointer[P, E]) Append(buf []byte, value P) []byte {
	return c.codec.Append(buf, (*E)(value))
}

func (c castPointer[P, E]) Put(buf []byte, value P) int {
	return c.codec.Put(buf, (*E)(value))
}

func (c castPointer[P, E]) Get(buf []byte) (P, int) {
	return c.codec.Get(buf)
}

func (c castPointer[P, E]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c castPointer[P, E]) nilsLast() Codec[P] {
	//nolint:forcetypeassert
	return castPointer[P, E]{c.codec.nilsLast().(pointerCodec[E])}
}

func (c castSlice[S, E]) Append(buf []byte, value S) []byte {
	return c.codec.Append(buf, []E(value))
}

func (c castSlice[S, E]) Put(buf []byte, value S) int {
	return c.codec.Put(buf, []E(value))
}

func (c castSlice[S, E]) Get(buf []byte) (S, int) {
	return c.codec.Get(buf)
}

func (c castSlice[S, E]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c castSlice[S, E]) nilsLast() Codec[S] {
	//nolint:forcetypeassert
	return castSlice[S, E]{c.codec.nilsLast().(sliceCodec[E])}
}

func (c castMap[M, K, V]) Append(buf []byte, value M) []byte {
	return c.codec.Append(buf, map[K]V(value))
}

func (c castMap[M, K, V]) Put(buf []byte, value M) int {
	return c.codec.Put(buf, map[K]V(value))
}

func (c castMap[M, K, V]) Get(buf []byte) (M, int) {
	return c.codec.Get(buf)
}

func (c castMap[M, K, V]) RequiresTerminator() bool {
	return c.codec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c castMap[M, K, V]) nilsLast() Codec[M] {
	//nolint:forcetypeassert
	return castMap[M, K, V]{c.codec.nilsLast().(mapCodec[K, V])}
}
