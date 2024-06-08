package lexy

import (
	"bytes"
	"io"
	"math/big"
	"time"
)

// Codec defines methods for encoding and decoding values to and from a binary
// form.
type Codec[T any] interface {
	// Unfortunately, a Codec can't be defined or created using
	// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler. Those types
	// require the value to be a receiver instead of an argument.

	// Write writes a value to the given io.Writer.
	Write(value T, w io.Writer) error

	// Read reads a value from the given io.Reader and returns it.
	Read(r io.Reader) (T, error)
}

// Encode uses codec to encode value into a []byte and returns it. This is a
// convenience function to create a new []byte.
func Encode[T any](codec Codec[T], value T) ([]byte, error) {
	var b bytes.Buffer
	if err := codec.Write(value, &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode uses codec to decode data into a value and returns it. This is a
// convenience function using a []byte.
func Decode[T any](codec Codec[T], data []byte) (T, error) {
	// TODO: NewBuffer takes ownership of data, this is probably a bad idea
	// here.
	return codec.Read(bytes.NewBuffer(data))
}

func BoolCodec() Codec[bool]                        { return boolCodec{} }
func UInt8Codec() Codec[uint8]                      { return uint8Codec{} }
func UInt16Codec() Codec[uint16]                    { return uint16Codec{} }
func UInt32Codec() Codec[uint32]                    { return uint32Codec{} }
func UInt64Codec() Codec[uint64]                    { return uint64Codec{} }
func Int8Codec() Codec[int8]                        { return int8Codec{} }
func Int16Codec() Codec[int16]                      { return int16Codec{} }
func Int32Codec() Codec[int32]                      { return int32Codec{} }
func Int64Codec() Codec[int64]                      { return int64Codec{} }
func Float32Codec() Codec[float32]                  { return float32Codec{} }
func Float64Codec() Codec[float64]                  { return float64Codec{} }
func BigIntCodec() Codec[big.Int]                   { return bigIntCodec{} }
func BigFloatCodec() Codec[big.Float]               { return bigFloatCodec{} }
func StringCodec() Codec[string]                    { return stringCodec{} }
func TimeCodec() Codec[time.Time]                   { return timeCodec{} }
func SliceCodec[T any]() Codec[[]T]                 { return sliceCodec[T]{} }
func MapCodec[K comparable, V any]() Codec[map[K]V] { return mapCodec[K, V]{} }
func StructCodec[T any]() Codec[T]                  { return structCodec[T]{} }

type boolCodec struct{}

func (b boolCodec) Read(r io.Reader) (bool, error) {
	panic("unimplemented")
}

func (b boolCodec) Write(value bool, w io.Writer) error {
	panic("unimplemented")
}

type uint8Codec struct{}

func (u uint8Codec) Read(r io.Reader) (uint8, error) {
	panic("unimplemented")
}

func (u uint8Codec) Write(value uint8, w io.Writer) error {
	panic("unimplemented")
}

type uint16Codec struct{}

func (u uint16Codec) Read(r io.Reader) (uint16, error) {
	panic("unimplemented")
}

func (u uint16Codec) Write(value uint16, w io.Writer) error {
	panic("unimplemented")
}

type uint32Codec struct{}

func (u uint32Codec) Read(r io.Reader) (uint32, error) {
	panic("unimplemented")
}

func (u uint32Codec) Write(value uint32, w io.Writer) error {
	panic("unimplemented")
}

type uint64Codec struct{}

func (u uint64Codec) Read(r io.Reader) (uint64, error) {
	panic("unimplemented")
}

func (u uint64Codec) Write(value uint64, w io.Writer) error {
	panic("unimplemented")
}

type int8Codec struct{}

func (i int8Codec) Read(r io.Reader) (int8, error) {
	panic("unimplemented")
}

func (i int8Codec) Write(value int8, w io.Writer) error {
	panic("unimplemented")
}

type int16Codec struct{}

func (i int16Codec) Read(r io.Reader) (int16, error) {
	panic("unimplemented")
}

func (i int16Codec) Write(value int16, w io.Writer) error {
	panic("unimplemented")
}

type int32Codec struct{}

func (i int32Codec) Read(r io.Reader) (int32, error) {
	panic("unimplemented")
}

func (i int32Codec) Write(value int32, w io.Writer) error {
	panic("unimplemented")
}

type int64Codec struct{}

func (i int64Codec) Read(r io.Reader) (int64, error) {
	panic("unimplemented")
}

func (i int64Codec) Write(value int64, w io.Writer) error {
	panic("unimplemented")
}

type float32Codec struct{}

func (f float32Codec) Read(r io.Reader) (float32, error) {
	panic("unimplemented")
}

func (f float32Codec) Write(value float32, w io.Writer) error {
	panic("unimplemented")
}

type float64Codec struct{}

func (f float64Codec) Read(r io.Reader) (float64, error) {
	panic("unimplemented")
}

func (f float64Codec) Write(value float64, w io.Writer) error {
	panic("unimplemented")
}

type bigIntCodec struct{}

func (b bigIntCodec) Read(r io.Reader) (big.Int, error) {
	panic("unimplemented")
}

func (b bigIntCodec) Write(value big.Int, w io.Writer) error {
	panic("unimplemented")
}

type bigFloatCodec struct{}

func (b bigFloatCodec) Read(r io.Reader) (big.Float, error) {
	panic("unimplemented")
}

func (b bigFloatCodec) Write(value big.Float, w io.Writer) error {
	panic("unimplemented")
}

type stringCodec struct{}

func (s stringCodec) Read(r io.Reader) (string, error) {
	panic("unimplemented")
}

func (s stringCodec) Write(value string, w io.Writer) error {
	panic("unimplemented")
}

type timeCodec struct{}

func (t timeCodec) Read(r io.Reader) (time.Time, error) {
	panic("unimplemented")
}

func (t timeCodec) Write(value time.Time, w io.Writer) error {
	panic("unimplemented")
}

type sliceCodec[T any] struct{}

func (s sliceCodec[T]) Read(r io.Reader) ([]T, error) {
	panic("unimplemented")
}

func (s sliceCodec[T]) Write(value []T, w io.Writer) error {
	panic("unimplemented")
}

type mapCodec[K comparable, V any] struct{}

func (m mapCodec[K, V]) Read(r io.Reader) (map[K]V, error) {
	panic("unimplemented")
}

func (m mapCodec[K, V]) Write(value map[K]V, w io.Writer) error {
	panic("unimplemented")
}

type structCodec[T any] struct{}

func (s structCodec[T]) Read(r io.Reader) (T, error) {
	panic("unimplemented")
}

func (s structCodec[T]) Write(value T, w io.Writer) error {
	panic("unimplemented")
}

// Decouple type prefixes, those are a feature of the aggregate Codec. The int8,
// int16, ... Codecs should not have prefixes.
