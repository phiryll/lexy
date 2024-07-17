package internal

import (
	"io"
	"math/big"
)

var (
	BigIntCodec   Codec[*big.Int]   = bigIntCodec{}
	BigFloatCodec Codec[*big.Float] = bigFloatCodec{}
)

// bigIntCodec is the Codec for big.Int values.
//
// Values are encoded using this logic:
//
//	b := value.Bytes() // absolute value as a big-endian byte slice
//	size := len(b)
//	if value < 0 {
//		write -size using Int64Codec
//		write b with all bits flipped
//	} else {
//		write +size using Int64Codec
//		write b
//	}
//
// This makes size (negative for negative values) the primary sort key,
// and the big-endian bytes for the value (bits flipped for negative values) the secondary sort key.
// The effect is that longer numbers will be ordered closer to +/-infinity.
// This works because bigInt.Bytes() will never have a leading zero byte.
type bigIntCodec struct{}

func (c bigIntCodec) Read(r io.Reader) (*big.Int, error) {
	neg := false
	size, err := int64Codec.Read(r)
	if err != nil {
		return nil, err
	}
	if size < 0 {
		neg = true
		size = -size
		// r is only used to read the value bits at this point,
		// so we can reassign it safely.
		r = negateReader{r}
	}
	b := make([]byte, size)
	n, err := r.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err == io.EOF {
		if int64(n) < size {
			return nil, io.ErrUnexpectedEOF
		}
		err = nil
	}
	var value big.Int
	value.SetBytes(b)
	if neg {
		value.Neg(&value)
	}
	return &value, nil
}

func (c bigIntCodec) Write(w io.Writer, value *big.Int) error {
	neg := false
	sign := value.Sign()
	b := value.Bytes()
	size := len(b)
	if sign < 0 {
		size = -size
		neg = true
	}
	if err := int64Codec.Write(w, int64(size)); err != nil {
		return err
	}
	if neg {
		w = negateWriter{w}
	}
	_, err := w.Write(b)
	return err
}

func (c bigIntCodec) RequiresTerminator() bool {
	return false
}

// bigFloatCodec is the Codec for big.Float values.
//
// This is roughly similar to the float32/64 Codecs, but there are some wrinkles.
// There is no good way to get the mantissa in a binary form,
// the standard library just doesn't expose that information.
// A description of the encoding and why it does what it does follows.
//
// Shift a copy of the big.Float so that:
//
//	all significant bits are to the left of the point,
//	the high bit of the high byte is 1, and
//	the low byte is not 0
//
// Get the big.Int value of the shifted big.Float.
// This is the mantissa if you interpret it as being immediately to the right of the point,
// which is the standard for representing a mantissa, 0.5 <= mantissa < 1.0.
// The only purpose of this is to get the exact []byte of the mantissa
// out of a big.Float without resorting to parsing.
// None of this will change the exponent or precision that is encoded.
//
// For example (binary, non-significant bits are shown as "-", assume they're all 0):
//
//	A = 7.0 (prec 3)  = 0.111- ----           * 2**3
//	B = 7.0 (prec 4)  = 0.1110 ----           * 2**3
//	C = 7.0 (prec 10) = 0.1110 0000 00-- ---- * 2**3
//
// All of these are the same semantic value, but with different precisions.
// After the shift we have (precision does not change)
//
//	A shift by 5  = 0.111- ----           * 2**8
//		prec = 3, prec - exp = 0
//	    Int mant  = 111- ---- = 224
//	B shift by 5  = 0.1110 ----           * 2**8
//		prec = 4, prec - exp = 1
//	    Int mant  = 1110 ---- = 224
//	C shift by 13 = 0.1110 0000 00-- ---- * 2**16
//		prec = 10, prec - exp = 7
//	    Int mant  = 1110 0000 00-- ---- = 57344
//
// Since the mantissa is variable length, it must be escaped and terminated.
// The precision and rounding mode must be encoded following
// the sign, exponent, and mantissa to keep the lexicographical ordering correct.
//
// We can see C > A and C > B since that's what the necessary encoding does.
// Therefore, B > A if the ordering is consistent with C > A and C > B, higher precisions are greater.
// For negative values, higher precisions are lesser.
// This leads to encoding the precision immediately after the mantissa.
//
// Encode:
//
//	write prefix int8: -3/-2/-1/+1/+2/+3 for
//		-Inf / (-Inf,0) / -0 / +0 / (0,+Inf) / +Inf
//	if infinite or zero, we're done
//	write int32 exponent
//		negate exponent first if Float is negative
//	write the (big-endian) bytes of the big.Int of the shifted mantissa
//		do *not* encode with bitIntCodec, write the raw bytes
//		trailing non-sigificant bits will already be zero, the conversion to big.Int requires it
//		escape and terminate, then flip bits (including the terminator) if Float is negative
//	write int32 precision
//		negate precision first if Float is negative
//	write uint8 rounding mode
type bigFloatCodec struct{}

const (
	negInf       int8 = -3
	negFinite    int8 = -2
	negZero      int8 = -1
	posZero      int8 = +1
	nonNegFinite int8 = +2
	posInf       int8 = +3
)

var modeCodec = uintCodec[big.RoundingMode]{}

func computeShift(exp int32, prec int32) int {
	// (prec - exp) is a shift of significant bits to immediately left of the point.
	shift := prec - exp
	// Shift a little further so the high bit in the high byte is 1.
	// Equivalently, the exponent is a multiple of 8.
	// There are exactly prec bits to that leading bit,
	// so shift enough to round up prec to the nearest multiple of 8.
	adjustment := (-prec) % 8
	if adjustment < 0 {
		adjustment += 8
	}
	return int(shift + adjustment)
}

func (c bigFloatCodec) Read(r io.Reader) (*big.Float, error) {
	prefix, err := int8Codec.Read(r)
	if err != nil {
		return nil, err
	}
	signbit := prefix < 0
	if prefix == negInf || prefix == posInf {
		var value big.Float
		return value.SetInf(signbit), nil
	}
	if prefix == negZero || prefix == posZero {
		var value big.Float
		if signbit {
			value.Neg(&value)
		}
		return &value, nil
	}
	mantReader := r
	if signbit {
		mantReader = negateReader{r}
	}

	exp, err := int32Codec.Read(r)
	if err != nil {
		return nil, err
	}
	mantBytes, err := unescape(mantReader)
	if err != nil {
		return nil, err
	}
	prec, err := int32Codec.Read(r)
	if err != nil {
		return nil, err
	}
	mode, err := modeCodec.Read(r)
	if err != nil {
		return nil, err
	}

	if signbit {
		exp = -exp
		prec = -prec
	}
	shift := computeShift(exp, prec)

	var mantInt big.Int
	var value big.Float
	mantInt.SetBytes(mantBytes)
	value.SetInt(&mantInt)
	value.SetMantExp(&value, -shift)
	value.SetPrec(uint(prec))
	value.SetMode(mode)
	if signbit {
		value.Neg(&value)
	}
	return &value, nil
}

func (c bigFloatCodec) Write(w io.Writer, value *big.Float) error {
	// exp and prec are int and uint, but internally they're 32 bits
	// use a signed prec here because we're doing possibly negative calculations with it
	signbit := value.Signbit() // true if negative or negative zero
	exp := int32(value.MantExp(nil))
	prec := int32(value.Prec())
	mode := value.Mode() // uint8
	shift := computeShift(exp, prec)

	isInf := value.IsInf()
	isZero := prec == 0

	var prefix int8
	switch {
	case isInf && signbit:
		prefix = negInf
	case isInf && !signbit:
		prefix = posInf
	case isZero && signbit:
		prefix = negZero
	case isZero && !signbit:
		prefix = posZero
	case signbit:
		prefix = negFinite
	case !signbit:
		prefix = nonNegFinite
	}
	if err := int8Codec.Write(w, prefix); err != nil {
		return err
	}
	if isInf || isZero {
		return nil
	}

	mantWriter := w
	if signbit {
		// These values are no longer being used except to write them
		exp = -exp
		prec = -prec
		mantWriter = negateWriter{w}
	}

	if err := int32Codec.Write(w, exp); err != nil {
		return err
	}

	var copy big.Float
	copy.Copy(value)
	copy.SetMantExp(&copy, shift)
	mantInt, acc := copy.Int(nil)
	if acc != big.Exact {
		panic("unexpected failure while encoding big.Float")
	}
	mantBytes := mantInt.Bytes()
	if _, err := escape(mantWriter, mantBytes); err != nil {
		return err
	}

	if err := int32Codec.Write(w, prec); err != nil {
		return err
	}
	return modeCodec.Write(w, mode)
}

func (c bigFloatCodec) RequiresTerminator() bool {
	return true
}
