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
	buf = stdFloat32.Append(buf, real(value))
	return stdFloat32.Append(buf, imag(value))
}

func (complex64Codec) Put(buf []byte, value complex64) int {
	n := stdFloat32.Put(buf, real(value))
	n += stdFloat32.Put(buf[n:], imag(value))
	return n
}

func (complex64Codec) Get(buf []byte) (complex64, int) {
	if len(buf) == 0 {
		return complex(0.0, 0.0), -1
	}
	n := 0
	realPart, count := stdFloat32.Get(buf)
	n += count
	if count < 0 {
		panic(io.ErrUnexpectedEOF)
	}
	imagPart, count := stdFloat32.Get(buf[n:])
	n += count
	if count < 0 {
		panic(io.ErrUnexpectedEOF)
	}
	return complex(realPart, imagPart), n
}

func (complex64Codec) RequiresTerminator() bool {
	return false
}

func (complex128Codec) Append(buf []byte, value complex128) []byte {
	buf = stdFloat64.Append(buf, real(value))
	return stdFloat64.Append(buf, imag(value))
}

func (complex128Codec) Put(buf []byte, value complex128) int {
	n := stdFloat64.Put(buf, real(value))
	n += stdFloat64.Put(buf[n:], imag(value))
	return n
}

func (complex128Codec) Get(buf []byte) (complex128, int) {
	if len(buf) == 0 {
		return complex(0.0, 0.0), -1
	}
	n := 0
	realPart, count := stdFloat64.Get(buf)
	n += count
	if count < 0 {
		panic(io.ErrUnexpectedEOF)
	}
	imagPart, count := stdFloat64.Get(buf[n:])
	n += count
	if count < 0 {
		panic(io.ErrUnexpectedEOF)
	}
	return complex(realPart, imagPart), n
}

func (complex128Codec) RequiresTerminator() bool {
	return false
}
