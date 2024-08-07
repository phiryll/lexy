package lexy_test

import (
	"io"
	"math"
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestNegateInt32(t *testing.T) {
	codec := lexy.Negate(lexy.Int32())
	testCodecRoundTrip(t, codec, []testCase[int32]{
		{"min", math.MinInt32, nil},
		{"-1", -1, nil},
		{"0", 0, nil},
		{"+1", 1, nil},
		{"max", math.MaxInt32, nil},
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
	encode := encoderFor(lexy.Negate(lexy.String()))
	assert.Less(t, encode("ab"), encode("a"))
}

func TestNegatePtrString(t *testing.T) {
	codec := lexy.Negate(toCodec(lexy.PointerTo(lexy.String())))
	testCodecRoundTrip(t, codec, []testCase[*string]{
		{"nil", nil, nil},
		{"*empty", ptr(""), nil},
		{"*abc", ptr("abc"), nil},
		{"*def", ptr("def"), nil},
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
	codec := lexy.Negate(toCodec(lexy.SliceOf(toCodec(lexy.PointerTo(lexy.String())))))

	testCodecRoundTrip(t, codec, []testCase[[]*string]{
		{"nil", nil, nil},
		{"[]", []*string{}, nil},
		{"[nil]", []*string{nil}, nil},
		{"*a", []*string{ptr("a")}, nil},
		{"*a, nil, *\"\", *xyz", []*string{ptr("a"), nil, ptr(""), ptr("xyz")}, nil},
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
	uint8  uint8
	pInt16 *int16
	string string
}

// order is: uint8, neg(string), neg(pInt16)
// putting the negated varying length field in the middle is intentional
type negateTestCodec struct{}

var (
	negPtrIntCodec = lexy.Negate(toCodec(lexy.PointerTo(lexy.Int16())))
	negStringCodec = lexy.Negate(lexy.String())

	negTestCodec lexy.Codec[negateTest] = negateTestCodec{}
)

func (n negateTestCodec) Read(r io.Reader) (negateTest, error) {
	var zero negateTest
	u8, err := lexy.Uint8().Read(r)
	if err != nil {
		return zero, err
	}
	s, err := negStringCodec.Read(r)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return zero, err
	}
	pInt, err := negPtrIntCodec.Read(r)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return zero, err
	}
	return negateTest{u8, pInt, s}, nil
}

func (n negateTestCodec) Write(w io.Writer, value negateTest) error {
	if err := lexy.Uint8().Write(w, value.uint8); err != nil {
		return err
	}
	if err := negStringCodec.Write(w, value.string); err != nil {
		return err
	}
	return negPtrIntCodec.Write(w, value.pInt16)
}

func (n negateTestCodec) RequiresTerminator() bool {
	return false
}

func TestNegateComplex(t *testing.T) {
	encode := encoderFor(negTestCodec)
	ptr := func(x int) *int16 {
		i16 := int16(x)
		return &i16
	}
	testCodecRoundTrip(t, negTestCodec, []testCase[negateTest]{
		{"{5, &100, def}", negateTest{5, ptr(100), "def"}, nil},
		{"{5, nil, \"\"}", negateTest{5, nil, ""}, nil},
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
