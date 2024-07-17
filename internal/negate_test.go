package internal_test

import (
	"io"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func TestNegateInt32(t *testing.T) {
	codec := internal.MakeNegateCodec(int32Codec)
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

func TestNegatePtrString(t *testing.T) {
	ptrCodec := internal.MakePointerCodec[*string](stringCodec)
	codec := internal.MakeNegateCodec(ptrCodec)
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
		encode(ptr("")),
		encode(nil),
	})
}

type negateTest struct {
	uint8  uint8
	pInt16 *int16
	string string
}

// order is [uint8, neg(pInt16), string]
type negateTestCodec struct{}

var negPIntCodec = internal.MakeNegateCodec(internal.MakePointerCodec[*int16](int16Codec))

func (n negateTestCodec) Read(r io.Reader) (negateTest, error) {
	var zero negateTest
	u8, err := uint8Codec.Read(r)
	if err != nil {
		return zero, err
	}
	pInt, err := negPIntCodec.Read(r)
	if err != nil {
		return zero, err
	}
	s, err := stringCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return negateTest{u8, pInt, s}, nil
}

func (n negateTestCodec) Write(w io.Writer, value negateTest) error {
	if err := uint8Codec.Write(w, value.uint8); err != nil {
		return err
	}
	if err := negPIntCodec.Write(w, value.pInt16); err != nil {
		return err
	}
	return stringCodec.Write(w, value.string)
}

func (n negateTestCodec) RequiresTerminator() bool {
	return false
}

func TestNegateComplex(t *testing.T) {
	encode := encoderFor(negateTestCodec{})
	ptr := func(x int) *int16 {
		i16 := int16(x)
		return &i16
	}
	assert.IsIncreasing(t, [][]byte{
		// sort order is [first, neg(second), third]
		encode(negateTest{5, ptr(100), ""}),
		encode(negateTest{5, ptr(100), "abc"}),
		encode(negateTest{5, ptr(100), "def"}),
		encode(negateTest{5, ptr(0), ""}),
		encode(negateTest{5, ptr(0), "abc"}),
		encode(negateTest{5, ptr(0), "def"}),
		encode(negateTest{5, ptr(-1), ""}),
		encode(negateTest{5, ptr(-1), "abc"}),
		encode(negateTest{5, ptr(-1), "def"}),
		encode(negateTest{5, ptr(-100), ""}),
		encode(negateTest{5, ptr(-100), "abc"}),
		encode(negateTest{5, ptr(-100), "def"}),
		encode(negateTest{5, nil, ""}),
		encode(negateTest{5, nil, "abc"}),
		encode(negateTest{5, nil, "def"}),

		encode(negateTest{10, ptr(100), ""}),
		encode(negateTest{10, ptr(100), "abc"}),
		encode(negateTest{10, ptr(100), "def"}),
		encode(negateTest{10, ptr(0), ""}),
		encode(negateTest{10, ptr(0), "abc"}),
		encode(negateTest{10, ptr(0), "def"}),
		encode(negateTest{10, ptr(-1), ""}),
		encode(negateTest{10, ptr(-1), "abc"}),
		encode(negateTest{10, ptr(-1), "def"}),
		encode(negateTest{10, ptr(-100), ""}),
		encode(negateTest{10, ptr(-100), "abc"}),
		encode(negateTest{10, ptr(-100), "def"}),
		encode(negateTest{10, nil, ""}),
		encode(negateTest{10, nil, "abc"}),
		encode(negateTest{10, nil, "def"})})
}
