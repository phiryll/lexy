package lexy

import (
	"io"
)

// complex64Codec is the Codec for complex64.
//
// The encoded order is real part first, imaginary part second.
type complex64Codec struct{}

func (complex64Codec) Read(r io.Reader) (complex64, error) {
	realPart, err := stdFloat32.Read(r)
	if err != nil {
		return 0, err
	}
	imagPart, err := stdFloat32.Read(r)
	if err != nil {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(realPart, imagPart), nil
}

func (complex64Codec) Write(w io.Writer, value complex64) error {
	if err := stdFloat32.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat32.Write(w, imag(value))
}

func (complex64Codec) RequiresTerminator() bool {
	return false
}

// complex128Codec is the Codec for complex128.
//
// The encoded order is real part first, imaginary part second.
type complex128Codec struct{}

func (complex128Codec) Read(r io.Reader) (complex128, error) {
	realPart, err := stdFloat64.Read(r)
	if err != nil {
		return 0, err
	}
	imagPart, err := stdFloat64.Read(r)
	if err != nil {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(realPart, imagPart), nil
}

func (complex128Codec) Write(w io.Writer, value complex128) error {
	if err := stdFloat64.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat64.Write(w, imag(value))
}

func (complex128Codec) RequiresTerminator() bool {
	return false
}
