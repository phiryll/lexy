package internal

import "io"

type Uint8Codec struct{}

func (c Uint8Codec) Read(r io.Reader) (uint8, error) {
	panic("unimplemented")
}

func (c Uint8Codec) Write(value uint8, w io.Writer) error {
	panic("unimplemented")
}

type Uint16Codec struct{}

func (c Uint16Codec) Read(r io.Reader) (uint16, error) {
	panic("unimplemented")
}

func (c Uint16Codec) Write(value uint16, w io.Writer) error {
	panic("unimplemented")
}

type Uint32Codec struct{}

func (c Uint32Codec) Read(r io.Reader) (uint32, error) {
	panic("unimplemented")
}

func (c Uint32Codec) Write(value uint32, w io.Writer) error {
	panic("unimplemented")
}

type Uint64Codec struct{}

func (c Uint64Codec) Read(r io.Reader) (uint64, error) {
	panic("unimplemented")
}

func (c Uint64Codec) Write(value uint64, w io.Writer) error {
	panic("unimplemented")
}

type Int8Codec struct{}

func (c Int8Codec) Read(r io.Reader) (int8, error) {
	panic("unimplemented")
}

func (c Int8Codec) Write(value int8, w io.Writer) error {
	panic("unimplemented")
}

type Int16Codec struct{}

func (c Int16Codec) Read(r io.Reader) (int16, error) {
	panic("unimplemented")
}

func (c Int16Codec) Write(value int16, w io.Writer) error {
	panic("unimplemented")
}

type Int32Codec struct{}

func (c Int32Codec) Read(r io.Reader) (int32, error) {
	panic("unimplemented")
}

func (c Int32Codec) Write(value int32, w io.Writer) error {
	panic("unimplemented")
}

type Int64Codec struct{}

func (c Int64Codec) Read(r io.Reader) (int64, error) {
	panic("unimplemented")
}

func (c Int64Codec) Write(value int64, w io.Writer) error {
	panic("unimplemented")
}
