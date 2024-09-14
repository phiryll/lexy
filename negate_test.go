package lexy_test

import (
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
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[int32]{
		{"min", math.MinInt32, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"-1", -1, []byte{0x80, 0x00, 0x00, 0x00}},
		{"0", 0, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
		{"+1", 1, []byte{0x7F, 0xFF, 0xFF, 0xFE}},
		{"max", math.MaxInt32, []byte{0x00, 0x00, 0x00, 0x00}},
	})
}

func TestNegateInt32Ordering(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.Int32())
	assert.False(t, codec.RequiresTerminator())
	testOrdering(t, codec, []testCase[int32]{
		{"max", math.MaxInt32, nil},
		{"100", 100, nil},
		{"1", 1, nil},
		{"0", 0, nil},
		{"-1", -1, nil},
		{"-100", -100, nil},
		{"min", math.MinInt32, nil},
	})
}

// The simple implementation is to simply invert all the bits, but it doesn't work.
// This tests for that regression, see the comments on negateEscapeCodec for details.
func TestNegateLength(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.String())
	assert.False(t, codec.RequiresTerminator())
	assert.Less(t, codec.Append(nil, "ab"), codec.Append(nil, "a"))
}

func TestNegatePtrString(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.PointerTo(lexy.String()))
	assert.False(t, codec.RequiresTerminator())
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
}

func TestNegatePtrStringOrdering(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.PointerTo(lexy.String()))
	assert.False(t, codec.RequiresTerminator())
	testOrdering(t, codec, []testCase[*string]{
		{"*def", ptr("def"), nil},
		{"*abc", ptr("abc"), nil},
		{"*ab", ptr("ab"), nil},
		{"*empty", ptr(""), nil},
		{"nil", nil, nil},
	})
}

func TestNegateSlicePtrString(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.SliceOf(lexy.PointerTo(lexy.String())))
	assert.False(t, codec.RequiresTerminator())
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
}

func TestNegateSlicePtrStringOrdering(t *testing.T) {
	t.Parallel()
	codec := lexy.Negate(lexy.SliceOf(lexy.PointerTo(lexy.String())))
	assert.False(t, codec.RequiresTerminator())
	testOrdering(t, codec, []testCase[[]*string]{
		{"[*b, nil]", []*string{ptr("b"), nil}, nil},
		{"[*b]", []*string{ptr("b")}, nil},
		{"[*a, *a]", []*string{ptr("a"), ptr("a")}, nil},
		{"[*a, *empty]", []*string{ptr("a"), ptr("")}, nil},
		{"[*a, nil, *z]", []*string{ptr("a"), nil, ptr("z")}, nil},
		{"[*a, nil, nil, nil, nil]", []*string{ptr("a"), nil, nil, nil, nil}, nil},
		{"[*a, nil]", []*string{ptr("a"), nil}, nil},
		{"[*a]", []*string{ptr("a")}, nil},
		{"[nil]", []*string{nil}, nil},
		{"[]", []*string{}, nil},
		{"nil", nil, nil},
	})
}

type negateTest struct {
	fUint8    uint8
	fPtrInt16 *int16
	fString   string
}

var (
	negPtrIntCodec = lexy.Negate(lexy.PointerTo(lexy.Int16()))
	negStringCodec = lexy.Negate(lexy.String())

	negTestCodec lexy.Codec[negateTest] = negateTestCodec{}
)

// Sort order is: uint8, neg(string), neg(pInt16).
// Putting the negated varying length field in the middle is intentional.
type negateTestCodec struct{}

func (negateTestCodec) Append(buf []byte, value negateTest) []byte {
	buf = lexy.Uint8().Append(buf, value.fUint8)
	buf = negStringCodec.Append(buf, value.fString)
	return negPtrIntCodec.Append(buf, value.fPtrInt16)
}

func (negateTestCodec) Put(buf []byte, value negateTest) []byte {
	buf = lexy.Uint8().Put(buf, value.fUint8)
	buf = negStringCodec.Put(buf, value.fString)
	return negPtrIntCodec.Put(buf, value.fPtrInt16)
}

func (negateTestCodec) Get(buf []byte) (negateTest, []byte) {
	u8, buf := lexy.Uint8().Get(buf)
	s, buf := negStringCodec.Get(buf)
	pInt, buf := negPtrIntCodec.Get(buf)
	return negateTest{u8, pInt, s}, buf
}

func (negateTestCodec) RequiresTerminator() bool {
	return false
}

func TestNegateComplex(t *testing.T) {
	t.Parallel()
	assert.False(t, negTestCodec.RequiresTerminator())
	testCodec(t, negTestCodec, []testCase[negateTest]{
		{"{5, &100, def}", negateTest{5, ptr(int16(100)), "def"}, concat(
			[]byte{0x05},
			negString("def"), []byte{negTerm},
			[]byte{negPNonNil, ^byte(0x80), ^byte(0x64)},
		)},
		{"{5, nil, \"\"}", negateTest{5, nil, ""}, []byte{
			0x05,
			negTerm,
			negPNilFirst,
		}},
	})
}

func TestNegateComplexOrdering(t *testing.T) {
	t.Parallel()
	p := ptr[int16]
	testOrdering(t, negTestCodec, []testCase[negateTest]{
		// sort order is: first, neg(third), neg(second)
		{"{5, *100, def}", negateTest{5, p(100), "def"}, nil},
		{"{5, *0, def}", negateTest{5, p(0), "def"}, nil},
		{"{5, *-1, def}", negateTest{5, p(-1), "def"}, nil},
		{"{5, *-100, def}", negateTest{5, p(-100), "def"}, nil},
		{"{5, nil, def}", negateTest{5, nil, "def"}, nil},

		{"{5, *100, abc}", negateTest{5, p(100), "abc"}, nil},
		{"{5, *0, abc}", negateTest{5, p(0), "abc"}, nil},
		{"{5, *-1, abc}", negateTest{5, p(-1), "abc"}, nil},
		{"{5, *-100, abc}", negateTest{5, p(-100), "abc"}, nil},
		{"{5, nil, abc}", negateTest{5, nil, "abc"}, nil},

		{"{5, *100, empty}", negateTest{5, p(100), ""}, nil},
		{"{5, *0, empty}", negateTest{5, p(0), ""}, nil},
		{"{5, *-1, empty}", negateTest{5, p(-1), ""}, nil},
		{"{5, *-100, empty}", negateTest{5, p(-100), ""}, nil},
		{"{5, nil, empty}", negateTest{5, nil, ""}, nil},

		{"{10, *100, def}", negateTest{10, p(100), "def"}, nil},
		{"{10, *0, def}", negateTest{10, p(0), "def"}, nil},
		{"{10, *-1, def}", negateTest{10, p(-1), "def"}, nil},
		{"{10, *-100, def}", negateTest{10, p(-100), "def"}, nil},
		{"{10, nil, def}", negateTest{10, nil, "def"}, nil},

		{"{10, *100, abc}", negateTest{10, p(100), "abc"}, nil},
		{"{10, *0, abc}", negateTest{10, p(0), "abc"}, nil},
		{"{10, *-1, abc}", negateTest{10, p(-1), "abc"}, nil},
		{"{10, *-100, abc}", negateTest{10, p(-100), "abc"}, nil},
		{"{10, nil, abc}", negateTest{10, nil, "abc"}, nil},

		{"{10, *100, empty}", negateTest{10, p(100), ""}, nil},
		{"{10, *0, empty}", negateTest{10, p(0), ""}, nil},
		{"{10, *-1, empty}", negateTest{10, p(-1), ""}, nil},
		{"{10, *-100, empty}", negateTest{10, p(-100), ""}, nil},
		{"{10, nil, empty}", negateTest{10, nil, ""}, nil},
	})
}

func TestNegateUnwrapsTerminator(t *testing.T) {
	t.Parallel()
	assert.Equal(t,
		lexy.Negate(lexy.String()),
		lexy.Negate(lexy.TerminatedString()))
}
