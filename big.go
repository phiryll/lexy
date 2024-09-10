package lexy

import (
	"math/big"
)

// bigIntCodec is the Codec for *big.Int values.
//
// Values are encoded using this logic:
//
//	write prefixNilFirst/Last if value is nil and return immediately
//	write prefixNonNil
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
type bigIntCodec struct {
	prefix Prefix
}

//nolint:mnd
func (c bigIntCodec) Append(buf []byte, value *big.Int) []byte {
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	size := (value.BitLen() + 7) / 8
	// Preallocate and put into it so we can use FillBytes, avoiding a copy.
	start := len(buf)
	buf = append(buf, make([]byte, size+8)...)
	putBuf := buf[start:]
	sign := value.Sign()
	if sign < 0 {
		putBuf = stdInt64.Put(putBuf, -int64(size))
		value.FillBytes(putBuf[:size])
		negate(putBuf[:size])
	} else {
		putBuf = stdInt64.Put(putBuf, int64(size))
		value.FillBytes(putBuf[:size])
	}
	return buf
}

//nolint:mnd
func (c bigIntCodec) Put(buf []byte, value *big.Int) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	size := (value.BitLen() + 7) / 8
	_ = buf[size+8-1] // check that we have room
	sign := value.Sign()
	if sign < 0 {
		buf = stdInt64.Put(buf, -int64(size))
		value.FillBytes(buf[:size])
		negate(buf[:size])
	} else {
		buf = stdInt64.Put(buf, int64(size))
		value.FillBytes(buf[:size])
	}
	return buf[size:]
}

func (c bigIntCodec) Get(buf []byte) (*big.Int, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	size, buf := stdInt64.Get(buf)
	var value big.Int
	if size == 0 {
		return &value, buf
	}
	if size < 0 {
		size = -size
		_ = buf[size-1]
		value.SetBytes(negCopy(buf[:size]))
		value.Neg(&value)
	} else {
		_ = buf[size-1]
		value.SetBytes(buf[:size])
	}
	return &value, buf[size:]
}

func (bigIntCodec) RequiresTerminator() bool {
	// One can't be a prefix of another because they would need to have the same first int64,
	// which is the number of bytes in the rest of the encoded value.
	return false
}

//lint:ignore U1000 this is actually used
func (bigIntCodec) nilsLast() Codec[*big.Int] {
	return bigIntCodec{PrefixNilsLast}
}

// bigFloatCodec is the Codec for *big.Float values.
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
// This is the mantissa if interpreted as being immediately to the right of the point,
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
//	write prefixNilFirst/Last if value is nil and return immediately
//	write prefixNonNil
//	write int8: negInf/negFinite/negZero/posZero/posFinite/posInf
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
type bigFloatCodec struct {
	prefix Prefix
}

// The second byte written in the *big.Float encoding after the initial prefixNonNil byte if non-nil.
// The values were chosen so that negInf < negFinite < negZero < posZero < posFinite < posInf.
// Neither the encoded values for these constants nor their complements need to be escaped.
const (
	negInf    int8 = -3
	negFinite int8 = -2
	negZero   int8 = -1
	posZero   int8 = +1
	posFinite int8 = +2
	posInf    int8 = +3
)

var modeCodec = castUint8[big.RoundingMode]{}

func computeShift(exp, prec int32) int {
	// (prec - exp) is a shift of significant bits to immediately left of the point.
	shift := prec - exp
	// Shift a little further so the high bit in the high byte is 1.
	// Equivalently, the exponent is a multiple of 8.
	// There are exactly prec bits to that leading bit,
	// so shift enough to round up prec to the nearest multiple of 8.
	//nolint:mnd
	adjustment := (-prec) % 8
	if adjustment < 0 {
		adjustment += 8
	}
	return int(shift + adjustment)
}

//nolint:cyclop,funlen,mnd
func (c bigFloatCodec) Append(buf []byte, value *big.Float) []byte {
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	// exp and prec are int and uint, but internally they're 32 bits
	// use a signed prec here because we're doing possibly negative calculations with it
	signbit := value.Signbit() // true if negative or negative zero
	exp := int32(value.MantExp(nil))
	prec := int32(value.Prec())
	mode := value.Mode() // uint8
	shift := computeShift(exp, prec)

	isInf := value.IsInf()
	isZero := prec == 0

	var kind int8
	switch {
	case isInf && signbit:
		kind = negInf
	case isInf && !signbit:
		kind = posInf
	case isZero && signbit:
		kind = negZero
	case isZero && !signbit:
		kind = posZero
	case signbit:
		kind = negFinite
	case !signbit:
		kind = posFinite
	}
	if isInf || isZero {
		return stdInt8.Append(buf, kind)
	}

	var tmp big.Float
	tmp.SetMantExp(value, shift)
	mantInt, acc := tmp.Int(nil)
	if acc != big.Exact {
		panic(errBigFloatEncoding)
	}

	mantSize := (mantInt.BitLen() + 7) / 8
	start := len(buf)
	// 10 = 1 (kind) + 4 (exp) + 4 (prec) + 1 (mode)
	buf = append(buf, make([]byte, mantSize+10)...)
	putBuf := buf[start:]
	putBuf = stdInt8.Put(putBuf, kind)
	if signbit {
		putBuf = stdInt32.Put(putBuf, -exp)
		mantInt.FillBytes(putBuf[:mantSize])
		n := termNumAdded(putBuf[:mantSize])
		buf = append(buf, make([]byte, n)...)
		putBuf = buf[start+5:]
		negTerm(putBuf[:mantSize+n], n)
		putBuf = stdInt32.Put(putBuf[mantSize+n:], -prec)
	} else {
		putBuf = stdInt32.Put(putBuf, +exp)
		mantInt.FillBytes(putBuf[:mantSize])
		n := termNumAdded(putBuf[:mantSize])
		buf = append(buf, make([]byte, n)...)
		putBuf = buf[start+5:]
		term(putBuf[:mantSize+n], n)
		putBuf = stdInt32.Put(putBuf[mantSize+n:], prec)
	}
	modeCodec.Put(putBuf, mode)
	return buf
}

//nolint:cyclop,mnd
func (c bigFloatCodec) Put(buf []byte, value *big.Float) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	// exp and prec are int and uint, but internally they're 32 bits
	// use a signed prec here because we're doing possibly negative calculations with it
	signbit := value.Signbit() // true if negative or negative zero
	exp := int32(value.MantExp(nil))
	prec := int32(value.Prec())
	mode := value.Mode() // uint8
	shift := computeShift(exp, prec)

	isInf := value.IsInf()
	isZero := prec == 0

	var kind int8
	switch {
	case isInf && signbit:
		kind = negInf
	case isInf && !signbit:
		kind = posInf
	case isZero && signbit:
		kind = negZero
	case isZero && !signbit:
		kind = posZero
	case signbit:
		kind = negFinite
	case !signbit:
		kind = posFinite
	}
	buf = stdInt8.Put(buf, kind)
	if isInf || isZero {
		return buf
	}

	var tmp big.Float
	tmp.SetMantExp(value, shift)
	mantInt, acc := tmp.Int(nil)
	if acc != big.Exact {
		panic(errBigFloatEncoding)
	}

	mantSize := (mantInt.BitLen() + 7) / 8
	if signbit {
		buf = stdInt32.Put(buf, -exp)
		mantInt.FillBytes(buf[:mantSize])
		n := termNumAdded(buf[:mantSize])
		negTerm(buf[:mantSize+n], n)
		buf = stdInt32.Put(buf[mantSize+n:], -prec)
	} else {
		buf = stdInt32.Put(buf, +exp)
		mantInt.FillBytes(buf[:mantSize])
		n := termNumAdded(buf[:mantSize])
		term(buf[:mantSize+n], n)
		buf = stdInt32.Put(buf[mantSize+n:], prec)
	}
	return modeCodec.Put(buf, mode)
}

func (c bigFloatCodec) Get(buf []byte) (*big.Float, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}

	kind, buf := stdInt8.Get(buf)
	signbit := kind < 0
	if kind == negInf || kind == posInf {
		var value big.Float
		return value.SetInf(signbit), buf
	}
	if kind == negZero || kind == posZero {
		var value big.Float
		if signbit {
			value.Neg(&value)
		}
		return &value, buf
	}

	exp, buf := stdInt32.Get(buf)
	var mantBytes []byte
	if signbit {
		mantBytes, buf = negTermGet(buf)
	} else {
		mantBytes, buf = termGet(buf)
	}
	prec, buf := stdInt32.Get(buf)
	mode, buf := modeCodec.Get(buf)

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
	return &value, buf
}

func (bigFloatCodec) RequiresTerminator() bool {
	// All encoded parts are either fixed-length or escaped.
	return false
}

//lint:ignore U1000 this is actually used
func (bigFloatCodec) nilsLast() Codec[*big.Float] {
	return bigFloatCodec{PrefixNilsLast}
}

// bigRatCodec is the Codec for *big.Rat values.
// The denominator cannot be zero.
// Note that big.Rat will normalize the numerator and denominator to lowest terms, including 0/N to 0/1.
//
// Values are encoded using this logic:
//
//	write prefixNilFirst/Last if value is nil and return immediately
//	write prefixNonNil
//	write the numerator with bigIntCodec
//	write the denominator with bigIntCodec
type bigRatCodec struct {
	prefix Prefix
}

func (c bigRatCodec) Append(buf []byte, value *big.Rat) []byte {
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	buf = stdBigInt.Append(buf, value.Num())
	return stdBigInt.Append(buf, value.Denom())
}

func (c bigRatCodec) Put(buf []byte, value *big.Rat) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	buf = stdBigInt.Put(buf, value.Num())
	return stdBigInt.Put(buf, value.Denom())
}

func (c bigRatCodec) Get(buf []byte) (*big.Rat, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	num, buf := stdBigInt.Get(buf)
	denom, buf := stdBigInt.Get(buf)
	var value big.Rat
	return value.SetFrac(num, denom), buf
}

func (bigRatCodec) RequiresTerminator() bool {
	return false
}

//lint:ignore U1000 this is actually used
func (bigRatCodec) nilsLast() Codec[*big.Rat] {
	return bigRatCodec{PrefixNilsLast}
}
