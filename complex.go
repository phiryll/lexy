package lexy

import "io"

// Codecs for complex64 and complex128 types.
//
// The encoded order is real part first, imaginary part second.
type (
	complex64Codec  struct{}
	complex128Codec struct{}
)

func (complex64Codec) Append(buf []byte, value complex64) []byte {
	return stdFloat32.Append(
		stdFloat32.Append(buf, real(value)),
		imag(value))
}

func (complex64Codec) Put(buf []byte, value complex64) int {
	n := stdFloat32.Put(buf, real(value))
	return n + stdFloat32.Put(buf, imag(value))
}

func (complex64Codec) Write(w io.Writer, value complex64) error {
	if err := stdFloat32.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat32.Write(w, imag(value))
}

func (complex64Codec) Get(buf []byte) (complex64, int) {
	realPart, n1 := stdFloat32.Get(buf)
	imagPart, n2 := stdFloat32.Get(buf[n1:])
	return complex(realPart, imagPart), n1 + n2
}

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

func (complex64Codec) MaxSize() int {
	//nolint:mnd
	return 2 * sizeUint32
}

func (complex64Codec) RequiresTerminator() bool {
	return false
}

func (complex128Codec) Append(buf []byte, value complex128) []byte {
	return stdFloat64.Append(
		stdFloat64.Append(buf, real(value)),
		imag(value))
}

func (complex128Codec) Put(buf []byte, value complex128) int {
	n := stdFloat64.Put(buf, real(value))
	return n + stdFloat64.Put(buf, imag(value))
}

func (complex128Codec) Write(w io.Writer, value complex128) error {
	if err := stdFloat64.Write(w, real(value)); err != nil {
		return err
	}
	return stdFloat64.Write(w, imag(value))
}

func (complex128Codec) Get(buf []byte) (complex128, int) {
	realPart, n1 := stdFloat64.Get(buf)
	imagPart, n2 := stdFloat64.Get(buf[n1:])
	return complex(realPart, imagPart), n1 + n2
}

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

func (complex128Codec) MaxSize() int {
	//nolint:mnd
	return 2 * sizeUint64
}

func (complex128Codec) RequiresTerminator() bool {
	return false
}
