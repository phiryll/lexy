package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
)

// There's not much point in testing emptyCodec itself,
// but there are good reasons to test it in combination with other Codecs.
// In particular, it demonstrates why RequiresTerminator() must return true
// if the Codec might encode zero bytes on Write.
// These tests should also catch if any of the aggregate Codecs don't handle termination correctly.

type emptyStruct struct{}

type (
	mKey   map[emptyStruct]uint8
	mValue map[uint8]emptyStruct
)

var (
	empty           = emptyStruct{}
	emptyCodec      = lexy.Empty[emptyStruct]()
	keyEmptyCodec   = lexy.MakeMapOf[mKey](emptyCodec, lexy.Uint8())
	valueEmptyCodec = lexy.MakeMapOf[mValue](lexy.Uint8(), emptyCodec)
)

func TestPointerEmpty(t *testing.T) {
	testCodec(t, lexy.PointerTo(emptyCodec), []testCase[*emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(empty), []byte{pNonNil}},
	})
	testCodecFail(t, lexy.PointerTo(emptyCodec), nil)
}

func TestSliceEmpty(t *testing.T) {
	testCodec(t, lexy.SliceOf(emptyCodec), []testCase[[]emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", []emptyStruct{}, []byte{pNonNil}},
		{"[empty]", []emptyStruct{empty}, []byte{pNonNil, term}},
		{"[3x empty]", []emptyStruct{empty, empty, empty}, []byte{
			pNonNil, term, term, term,
		}},
	})
	testCodecFail(t, lexy.SliceOf(emptyCodec), nil)
}

func TestMapKeyEmpty(t *testing.T) {
	testCodec(t, keyEmptyCodec, []testCase[mKey]{
		{"nil", nil, []byte{pNilFirst}},
		{"{}", mKey{}, []byte{pNonNil}},
		{"{empty:2}", mKey{empty: 2}, []byte{
			pNonNil,
			term,
			0x02,
		}},
	})
	testCodecFail(t, keyEmptyCodec, nil)
}

func TestMapValueEmpty(t *testing.T) {
	testCodec(t, valueEmptyCodec, []testCase[mValue]{
		{"nil", nil, []byte{pNilFirst}},
		{"{}", mValue{}, []byte{pNonNil}},
		{"{2:empty}", mValue{2: empty}, []byte{
			pNonNil,
			0x02,
			term,
		}},
	})
	testCodecRoundTrip(t, valueEmptyCodec, []testCase[mValue]{
		{"non-trivial", mValue{
			1:   empty,
			167: empty,
			4:   empty,
			17:  empty,
		}, nil},
	})
	testCodecFail(t, valueEmptyCodec, nil)
}

func TestNegateEmpty(t *testing.T) {
	testCodec(t, lexy.Negate(emptyCodec), []testCase[emptyStruct]{
		{"neg(empty)", empty, []byte{0xFF}},
	})
	testCodecFail(t, lexy.Negate(emptyCodec), empty)
}

func TestTerminateEmpty(t *testing.T) {
	testCodec(t, lexy.Terminate(emptyCodec), []testCase[emptyStruct]{
		{"terminate(empty)", empty, []byte{0x00}},
	})
	testCodecFail(t, lexy.Terminate(emptyCodec), empty)
}
