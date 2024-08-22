package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
)

// There are good reasons to test emptyCodec in combination with other Codecs.
// In particular, it demonstrates why RequiresTerminator() must return true
// if the Codec might encode zero bytes on Write.
// These tests should also catch if any of the aggregate Codecs don't handle termination correctly.

type emptyStruct struct{}

type mValue map[uint8]emptyStruct

var (
	empty           = emptyStruct{}
	emptyCodec      = lexy.Empty[emptyStruct]()
	valueEmptyCodec = lexy.MakeMapOf[mValue](lexy.Uint8(), emptyCodec)
)

func TestEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, emptyCodec, []testCase[emptyStruct]{
		{"empty", emptyStruct{}, []byte{}},
	})
}

func TestPointerEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.PointerTo(emptyCodec), []testCase[*emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(empty), []byte{pNonNil}},
	})
}

func TestSliceEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.SliceOf(emptyCodec), []testCase[[]emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", []emptyStruct{}, []byte{pNonNil}},
		{"[empty]", []emptyStruct{empty}, []byte{pNonNil, term}},
		{"[3x empty]", []emptyStruct{empty, empty, empty}, []byte{
			pNonNil, term, term, term,
		}},
	})
}

func TestMapValueEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, valueEmptyCodec, []testCase[mValue]{
		{"nil", nil, []byte{pNilFirst}},
		{"{}", mValue{}, []byte{pNonNil}},
		{"{2:empty}", mValue{2: empty}, []byte{
			pNonNil,
			0x02,
			term,
		}},
	})
	testVaryingCodec(t, valueEmptyCodec, []testCase[mValue]{
		{"non-trivial", mValue{
			1:   empty,
			167: empty,
			4:   empty,
			17:  empty,
		}, nil},
	})
}

func TestNegateEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Negate(emptyCodec), []testCase[emptyStruct]{
		{"neg(empty)", empty, []byte{0xFF}},
	})
}

func TestTerminateEmpty(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.TerminateIfNeeded(emptyCodec), []testCase[emptyStruct]{
		{"terminate(empty)", empty, []byte{0x00}},
	})
}
