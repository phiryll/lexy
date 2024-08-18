package lexy_test

import (
	"io"
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

const (
	negTerm      = ^term      // 0xFF
	negEsc       = ^esc       // 0xFE
	negPNilFirst = ^pNilFirst // 0xFD
	negPNonNil   = ^pNonNil   // 0xFC
	negPNilLast  = ^pNilLast  // 0x02
)

// Assumes no 0s or 1s in the string that would need to be escaped.
func negString(s string) []byte {
	buf := make([]byte, len(s))
	for i, ch := range s {
		buf[i] = ^byte(ch)
	}
	return buf
}

func TestNegateInt32(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.Int32())
	testCodec(t, codec, []testCase[int32]{
		{"min", math.MinInt32, []byte{negEsc, 0xFF, negEsc, 0xFF, negEsc, 0xFF, negEsc, 0xFF, negTerm}},
		{"-1", -1, []byte{0x80, 0x00, 0x00, 0x00, negTerm}},
		{"0", 0, []byte{0x7F, negEsc, 0xFF, negEsc, 0xFF, negEsc, 0xFF, negTerm}},
		{"+1", 1, []byte{0x7F, negEsc, 0xFF, negEsc, 0xFF, negEsc, 0xFE, negTerm}},
		{"max", math.MaxInt32, []byte{0x00, 0x00, 0x00, 0x00, negTerm}},
	})

	encode := encoderFor(codec)
	assert.IsIncreasing(t, [][]byte{
		encode(math.MaxInt32),
		encode(100),
		encode(1),
		encode(0),
		encode(-1),
		encode(-100),
		encode(math.MinInt32),
	})
}

// The simple implementation is to simply invert all the bits, but it doesn't work.
// This tests for that regression, see the comments on negateCodec for details.
func TestNegateLength(t *testing.T) {
	t.Parallel()
	encode := encoderFor(lexy.Negate(lexy.String()))
	assert.Less(t, encode("ab"), encode("a"))
}

func TestNegatePtrString(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(toCodec(lexy.PointerTo(lexy.String())))
	testCodec(t, codec, []testCase[*string]{
		{"nil", nil, []byte{negPNilFirst, negTerm}},
		{"*empty", ptr(""), []byte{negPNonNil, negTerm}},
		{"*abc", ptr("abc"), concat(
			[]byte{negPNonNil},
			negString("abc"),
			[]byte{negTerm})},
		{"*def", ptr("def"), concat(
			[]byte{negPNonNil},
			negString("def"),
			[]byte{negTerm})},
	})

	encode := encoderFor(codec)
	assert.IsIncreasing(t, [][]byte{
		encode(ptr("def")),
		encode(ptr("abc")),
		encode(ptr("ab")),
		encode(ptr("")),
		encode(nil),
	})
}

func TestNegateSlicePtrString(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(toCodec(lexy.SliceOf(toCodec(lexy.PointerTo(lexy.String())))))
	// neg([]*string)
	// negate and slice codecs are escaping and terminating.
	testCodec(t, codec, []testCase[[]*string]{
		{"nil", nil, []byte{negPNilFirst, negTerm}},
		{"[]", []*string{}, []byte{negPNonNil, negTerm}},
		{"[nil]", []*string{nil}, []byte{negPNonNil, negPNilFirst, negEsc, negTerm, negTerm}},
		{"[*a]", []*string{ptr("a")}, concat(
			[]byte{negPNonNil},
			// esc(*a) by slice     => [non-nil, a, term]
			// ^(esc(..)) by negate => ^([non-nil, a, esc, term] ... term)
			[]byte{negPNonNil}, negString("a"), []byte{negEsc, negTerm},
			[]byte{negTerm})},
		{"[*a, nil, *\"\", *xyz]", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, concat(
			[]byte{negPNonNil},
			[]byte{negPNonNil}, negString("a"), []byte{negEsc, negTerm},
			[]byte{negPNilFirst, negEsc, negTerm},
			[]byte{negPNonNil}, negString(""), []byte{negEsc, negTerm},
			[]byte{negPNonNil}, negString("xyz"), []byte{negEsc, negTerm},
			[]byte{negTerm})},
	})

	encode := encoderFor(codec)
	assert.IsIncreasing(t, [][]byte{
		encode([]*string{ptr("b"), nil}),
		encode([]*string{ptr("b")}),
		encode([]*string{ptr("a"), ptr("a")}),
		encode([]*string{ptr("a"), ptr("")}),
		encode([]*string{ptr("a"), nil, ptr("z")}),
		encode([]*string{ptr("a"), nil, nil, nil, nil}),
		encode([]*string{ptr("a"), nil}),
		encode([]*string{ptr("a")}),
		encode([]*string{nil}),
		encode([]*string{}),
		encode(nil),
	})
}

type negateTest struct {
	fUint8    uint8
	fPtrInt16 *int16
	fString   string
}

var (
	negPtrIntCodec = lexy.Negate(toCodec(lexy.PointerTo(lexy.Int16())))
	negStringCodec = lexy.Negate(lexy.String())

	negTestCodec lexy.Codec[negateTest] = negateTestCodec{}
)

// Sort order is: uint8, neg(string), neg(pInt16).
// Putting the negated varying length field in the middle is intentional.
type negateTestCodec struct{}

func (c negateTestCodec) Append(buf []byte, value negateTest) []byte {
	return lexy.AppendUsingWrite[negateTest](c, buf, value)
}

func (c negateTestCodec) Put(buf []byte, value negateTest) int {
	return lexy.PutUsingAppend[negateTest](c, buf, value)
}

func (c negateTestCodec) Get(buf []byte) (negateTest, int) {
	return lexy.GetUsingRead[negateTest](c, buf)
}

func (negateTestCodec) Write(w io.Writer, value negateTest) error {
	if err := lexy.Uint8().Write(w, value.fUint8); err != nil {
		return err
	}
	if err := negStringCodec.Write(w, value.fString); err != nil {
		return err
	}
	return negPtrIntCodec.Write(w, value.fPtrInt16)
}

func (negateTestCodec) Read(r io.Reader) (negateTest, error) {
	var zero negateTest
	u8, err := lexy.Uint8().Read(r)
	if err != nil {
		return zero, err
	}
	s, err := negStringCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	pInt, err := negPtrIntCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return negateTest{u8, pInt, s}, nil
}

func (negateTestCodec) RequiresTerminator() bool {
	return false
}

func TestNegateComplex(t *testing.T) {
	t.Parallel()
	encode := encoderFor(negTestCodec)
	ptr := func(x int) *int16 {
		i16 := int16(x)
		return &i16
	}
	testCodec(t, negTestCodec, []testCase[negateTest]{
		{"{5, &100, def}", negateTest{5, ptr(100), "def"}, concat(
			[]byte{0x05},
			negString("def"), []byte{negTerm},
			[]byte{negPNonNil, ^byte(0x80), ^byte(0x64), negTerm},
		)},
		{"{5, nil, \"\"}", negateTest{5, nil, ""}, []byte{
			0x05,
			negTerm,
			negPNilFirst, negTerm,
		}},
	})

	assert.IsIncreasing(t, [][]byte{
		// sort order is: first, neg(third), neg(second)
		encode(negateTest{5, ptr(100), "def"}),
		encode(negateTest{5, ptr(0), "def"}),
		encode(negateTest{5, ptr(-1), "def"}),
		encode(negateTest{5, ptr(-100), "def"}),
		encode(negateTest{5, nil, "def"}),

		encode(negateTest{5, ptr(100), "abc"}),
		encode(negateTest{5, ptr(0), "abc"}),
		encode(negateTest{5, ptr(-1), "abc"}),
		encode(negateTest{5, ptr(-100), "abc"}),
		encode(negateTest{5, nil, "abc"}),

		encode(negateTest{5, ptr(100), ""}),
		encode(negateTest{5, ptr(0), ""}),
		encode(negateTest{5, ptr(-1), ""}),
		encode(negateTest{5, ptr(-100), ""}),
		encode(negateTest{5, nil, ""}),

		encode(negateTest{10, ptr(100), "def"}),
		encode(negateTest{10, ptr(0), "def"}),
		encode(negateTest{10, ptr(-1), "def"}),
		encode(negateTest{10, ptr(-100), "def"}),
		encode(negateTest{10, nil, "def"}),

		encode(negateTest{10, ptr(100), "abc"}),
		encode(negateTest{10, ptr(0), "abc"}),
		encode(negateTest{10, ptr(-1), "abc"}),
		encode(negateTest{10, ptr(-100), "abc"}),
		encode(negateTest{10, nil, "abc"}),

		encode(negateTest{10, ptr(100), ""}),
		encode(negateTest{10, ptr(0), ""}),
		encode(negateTest{10, ptr(-1), ""}),
		encode(negateTest{10, ptr(-100), ""}),
		encode(negateTest{10, nil, ""}),
	})
}
