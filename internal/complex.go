package internal

import "io"

var (
	Complex64Codec  Codec[complex64]  = complex64Codec{}
	Complex128Codec Codec[complex128] = complex128Codec{}
)

// complex64Codec is the Codec for complex64.
//
// The encoded order is real part first, imaginary part second.
type complex64Codec struct{}

func (c complex64Codec) Read(r io.Reader) (complex64, error) {
	real, err := basicFloat32Codec.Read(r)
	if err != nil {
		return 0, unexpectedIfEOF(err)
	}
	imag, err := basicFloat32Codec.Read(r)
	if err != nil && err != io.EOF {
		return 0, err
	}
	return complex(real, imag), nil
}

func (c complex64Codec) Write(w io.Writer, value complex64) error {
	if err := basicFloat32Codec.Write(w, real(value)); err != nil {
		return err
	}
	return basicFloat32Codec.Write(w, imag(value))
}

func (c complex64Codec) RequiresTerminator() bool {
	return false
}

// complex128Codec is the Codec for complex128.
//
// The encoded order is real part first, imaginary part second.
type complex128Codec struct{}

func (c complex128Codec) Read(r io.Reader) (complex128, error) {
	real, err := basicFloat64Codec.Read(r)
	if err != nil {
		return 0, unexpectedIfEOF(err)
	}
	imag, err := basicFloat64Codec.Read(r)
	if err != nil && err != io.EOF {
		return 0, err
	}
	return complex(real, imag), nil
}

func (c complex128Codec) Write(w io.Writer, value complex128) error {
	if err := basicFloat64Codec.Write(w, real(value)); err != nil {
		return err
	}
	return basicFloat64Codec.Write(w, imag(value))
}

func (c complex128Codec) RequiresTerminator() bool {
	return false
}
