package lexy

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

func (complex64Codec) Put(buf []byte, value complex64) []byte {
	buf = stdFloat32.Put(buf, real(value))
	return stdFloat32.Put(buf, imag(value))
}

func (complex64Codec) Get(buf []byte) (complex64, []byte) {
	realPart, buf := stdFloat32.Get(buf)
	imagPart, buf := stdFloat32.Get(buf)
	return complex(realPart, imagPart), buf
}

func (complex64Codec) RequiresTerminator() bool {
	return false
}

func (complex128Codec) Append(buf []byte, value complex128) []byte {
	buf = stdFloat64.Append(buf, real(value))
	return stdFloat64.Append(buf, imag(value))
}

func (complex128Codec) Put(buf []byte, value complex128) []byte {
	buf = stdFloat64.Put(buf, real(value))
	return stdFloat64.Put(buf, imag(value))
}

func (complex128Codec) Get(buf []byte) (complex128, []byte) {
	realPart, buf := stdFloat64.Get(buf)
	imagPart, buf := stdFloat64.Get(buf)
	return complex(realPart, imagPart), buf
}

func (complex128Codec) RequiresTerminator() bool {
	return false
}
