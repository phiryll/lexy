package lexy

import "io"

// complex64Codec is the Codec for complex64.
//
// The encoded order is real part first, imaginary part second.
type complex64Codec struct{}

func (c complex64Codec) Read(r io.Reader) (complex64, error) {
	realPart, err := stdFloat32Codec.Read(r)
	if err != nil {
		return 0, err
	}
	imagPart, err := stdFloat32Codec.Read(r)
	if err != nil {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(realPart, imagPart), nil
}

func (c complex64Codec) Write(w io.Writer, value complex64) error {
	if err := stdFloat32Codec.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat32Codec.Write(w, imag(value))
}

func (c complex64Codec) RequiresTerminator() bool {
	return false
}

// complex128Codec is the Codec for complex128.
//
// The encoded order is real part first, imaginary part second.
type complex128Codec struct{}

func (c complex128Codec) Read(r io.Reader) (complex128, error) {
	realPart, err := stdFloat64Codec.Read(r)
	if err != nil {
		return 0, err
	}
	imagPart, err := stdFloat64Codec.Read(r)
	if err != nil && err != io.EOF {
		return 0, UnexpectedIfEOF(err)
	}
	return complex(realPart, imagPart), nil
}

func (c complex128Codec) Write(w io.Writer, value complex128) error {
	if err := stdFloat64Codec.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat64Codec.Write(w, imag(value))
}

func (c complex128Codec) RequiresTerminator() bool {
	return false
}
