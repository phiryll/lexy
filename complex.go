package lexy

import "io"

// complex64Codec is the Codec for complex64.
//
// The encoded order is real part first, imaginary part second.
type complex64Codec struct{}

func (c complex64Codec) Read(r io.Reader) (complex64, error) {
	real, err := aFloat32Codec.Read(r)
	if err != nil {
		return 0, err
	}
	imag, err := aFloat32Codec.Read(r)
	if err != nil {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(real, imag), nil
}

func (c complex64Codec) Write(w io.Writer, value complex64) error {
	if err := aFloat32Codec.Write(w, real(value)); err != nil {
		return err
	}
	return aFloat32Codec.Write(w, imag(value))
}

func (c complex64Codec) RequiresTerminator() bool {
	return false
}

// complex128Codec is the Codec for complex128.
//
// The encoded order is real part first, imaginary part second.
type complex128Codec struct{}

func (c complex128Codec) Read(r io.Reader) (complex128, error) {
	real, err := aFloat64Codec.Read(r)
	if err != nil {
		return 0, err
	}
	imag, err := aFloat64Codec.Read(r)
	if err != nil && err != io.EOF {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(real, imag), nil
}

func (c complex128Codec) Write(w io.Writer, value complex128) error {
	if err := aFloat64Codec.Write(w, real(value)); err != nil {
		return err
	}
	return aFloat64Codec.Write(w, imag(value))
}

func (c complex128Codec) RequiresTerminator() bool {
	return false
}
