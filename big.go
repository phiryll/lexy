package lexy

import (
	"io"
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

func (c bigIntCodec) Append(buf []byte, value *big.Int) []byte {
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	sign := value.Sign()
	b := value.Bytes()
	size := int64(len(b))
	if sign < 0 {
		newBuf = stdInt64.Append(newBuf, -size)
		negate(b)
	} else {
		newBuf = stdInt64.Append(newBuf, size)
	}
	return append(newBuf, b...)
}

func (c bigIntCodec) Put(buf []byte, value *big.Int) int {
	// It would be nice to use big.Int.FillBytes to avoid an extra copy,
	// but it clears the entire buffer.
	// So it makes sense here to use Append.
	return mustCopy(buf, c.Append(nil, value))
}

func (c bigIntCodec) Get(buf []byte) (*big.Int, int) {
	// It's not efficient for Get and Read to share code,
	// because Read can negate its buffer directly if the value is negative,
	// while Get must make a copy first.
	if c.prefix.Get(buf) {
		return nil, 1
	}
	buf = buf[1:]
	size, n := stdInt64.Get(buf)
	buf = buf[n:]
	var value big.Int
	if size < 0 {
		size = -size
		value.SetBytes(negate(append([]byte(nil), buf[:size]...)))
		value.Neg(&value)
	} else {
		value.SetBytes(buf[:size])
	}
	return &value, 1 + n + int(size)
}

func (c bigIntCodec) Write(w io.Writer, value *big.Int) error {
	// The encoded bytes can't be written directly to w without
	// creating a temporary buffer holding most of them anyway,
	// so we might as well just reuse Append.
	_, err := w.Write(c.Append(nil, value))
	return err
}

func (c bigIntCodec) Read(r io.Reader) (*big.Int, error) {
	if done, err := c.prefix.Read(r); done {
		return nil, err
	}
	size, err := stdInt64.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	neg := false
	if size < 0 {
		neg = true
		size = -size
	}
	b := make([]byte, size)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	var value big.Int
	if neg {
		value.SetBytes(negate(b))
		value.Neg(&value)
	} else {
		value.SetBytes(b)
	}
	return &value, nil
}

func (bigIntCodec) RequiresTerminator() bool {
	return false
}

func (bigIntCodec) NilsLast() NillableCodec[*big.Int] {
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

func (c bigFloatCodec) Append(buf []byte, value *big.Float) []byte {
	return AppendUsingWrite[*big.Float](c, buf, value)
}

func (c bigFloatCodec) Put(buf []byte, value *big.Float) int {
	return PutUsingAppend[*big.Float](c, buf, value)
}

func (c bigFloatCodec) Get(buf []byte) (*big.Float, int) {
	return GetUsingRead[*big.Float](c, buf)
}

//nolint:cyclop,funlen
func (c bigFloatCodec) Write(w io.Writer, value *big.Float) error {
	if done, err := c.prefix.Write(w, value == nil); done {
		return err
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
	if err := stdInt8.Write(w, kind); err != nil {
		return err
	}
	if isInf || isZero {
		return nil
	}

	mantWriter := w
	if signbit {
		// These values are no longer being used except to write them.
		exp = -exp
		prec = -prec
		mantWriter = negateWriter{w}
	}

	if err := stdInt32.Write(w, exp); err != nil {
		return err
	}

	var tmp big.Float
	tmp.Copy(value)
	tmp.SetMantExp(&tmp, shift)
	mantInt, acc := tmp.Int(nil)
	if acc != big.Exact {
		panic(errBigFloatEncoding)
	}
	mantBytes := mantInt.Bytes()
	// order needs to be escape the bytes and *then* negate them if needed
	if _, err := doEscape(mantWriter, mantBytes); err != nil {
		return err
	}

	if err := stdInt32.Write(w, prec); err != nil {
		return err
	}
	return modeCodec.Write(w, mode)
}

//nolint:funlen
func (c bigFloatCodec) Read(r io.Reader) (*big.Float, error) {
	if done, err := c.prefix.Read(r); done {
		return nil, err
	}
	kind, err := stdInt8.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	signbit := kind < 0
	if kind == negInf || kind == posInf {
		var value big.Float
		return value.SetInf(signbit), nil
	}
	if kind == negZero || kind == posZero {
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

	exp, err := stdInt32.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	mantBytes, err := doUnescape(mantReader)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	prec, err := stdInt32.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	mode, err := modeCodec.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
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

func (bigFloatCodec) RequiresTerminator() bool {
	return true
}

func (bigFloatCodec) NilsLast() NillableCodec[*big.Float] {
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
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	newBuf = stdBigInt.Append(newBuf, value.Num())
	return stdBigInt.Append(newBuf, value.Denom())
}

func (c bigRatCodec) Put(buf []byte, value *big.Rat) int {
	if c.prefix.Put(buf, value == nil) {
		return 1
	}
	n := 1
	n += stdBigInt.Put(buf[n:], value.Num())
	return n + stdBigInt.Put(buf[n:], value.Denom())
}

func (c bigRatCodec) Get(buf []byte) (*big.Rat, int) {
	if c.prefix.Get(buf) {
		return nil, 1
	}
	num, nNum := stdBigInt.Get(buf[1:])
	denom, nDenom := stdBigInt.Get(buf[1+nNum:])
	var value big.Rat
	return value.SetFrac(num, denom), 1 + nNum + nDenom
}

func (c bigRatCodec) Write(w io.Writer, value *big.Rat) error {
	if done, err := c.prefix.Write(w, value == nil); done {
		return err
	}
	if err := stdBigInt.Write(w, value.Num()); err != nil {
		return err
	}
	return stdBigInt.Write(w, value.Denom())
}

func (c bigRatCodec) Read(r io.Reader) (*big.Rat, error) {
	if done, err := c.prefix.Read(r); done {
		return nil, err
	}
	num, err := stdBigInt.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	denom, err := stdBigInt.Read(r)
	if err != nil {
		return nil, UnexpectedIfEOF(err)
	}
	var value big.Rat
	return value.SetFrac(num, denom), nil
}

func (bigRatCodec) RequiresTerminator() bool {
	return false
}

func (bigRatCodec) NilsLast() NillableCodec[*big.Rat] {
	return bigRatCodec{PrefixNilsLast}
}
