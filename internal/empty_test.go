package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
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
	emptyCodec      = internal.EmptyCodec[emptyStruct]()
	ptrEmpty        = internal.PointerCodec[*emptyStruct](emptyCodec, true)
	arrayEmpty      = internal.ArrayCodec[[3]emptyStruct](emptyCodec)
	ptrToArrayEmpty = internal.PointerToArrayCodec[*[3]emptyStruct](emptyCodec, true)
	sliceEmpty      = internal.SliceCodec[[]emptyStruct](emptyCodec, true)
	mapKeyEmpty     = internal.MapCodec[mKey](emptyCodec, uint8Codec, true)
	mapValueEmpty   = internal.MapCodec[mValue](uint8Codec, emptyCodec, true)
	negateEmpty     = internal.NegateCodec(emptyCodec)
	terminateEmpty  = internal.Terminate(emptyCodec)
)

func TestPointerEmpty(t *testing.T) {
	testCodec(t, ptrEmpty, []testCase[*emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"*empty", ptr(empty), []byte{pNonNil}},
	})
	testCodecFail(t, ptrEmpty, nil)
}

func TestArrayEmpty(t *testing.T) {
	testCodec(t, arrayEmpty, []testCase[[3]emptyStruct]{
		{"[3x empty]", [3]emptyStruct{empty, empty, empty}, []byte{
			pNonNil, term, term, term,
		}},
	})
	testCodecFail(t, arrayEmpty, [3]emptyStruct{empty, empty, empty})
}

func TestPtrToArrayEmpty(t *testing.T) {
	testCodec(t, ptrToArrayEmpty, []testCase[*[3]emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"*[3x empty]", &[3]emptyStruct{empty, empty, empty}, []byte{
			pNonNil, term, term, term,
		}},
	})
	testCodecFail(t, ptrToArrayEmpty, nil)
}

func TestSliceEmpty(t *testing.T) {
	testCodec(t, sliceEmpty, []testCase[[]emptyStruct]{
		{"nil", nil, []byte{pNilFirst}},
		{"[]", []emptyStruct{}, []byte{pNonNil}},
		{"[empty]", []emptyStruct{empty}, []byte{pNonNil, term}},
		{"[3x empty]", []emptyStruct{empty, empty, empty}, []byte{
			pNonNil, term, term, term,
		}},
	})
	testCodecFail(t, sliceEmpty, nil)
}

func TestMapKeyEmpty(t *testing.T) {
	testCodec(t, mapKeyEmpty, []testCase[mKey]{
		{"nil", nil, []byte{pNilFirst}},
		{"{}", mKey{}, []byte{pNonNil}},
		{"{empty:2}", mKey{empty: 2}, []byte{
			pNonNil,
			term,
			0x02,
		}},
	})
	testCodecFail(t, mapKeyEmpty, nil)
}

func TestMapValueEmpty(t *testing.T) {
	testCodec(t, mapValueEmpty, []testCase[mValue]{
		{"nil", nil, []byte{pNilFirst}},
		{"{}", mValue{}, []byte{pNonNil}},
		{"{2:empty}", mValue{2: empty}, []byte{
			pNonNil,
			0x02,
			term,
		}},
	})
	testCodecRoundTrip(t, mapValueEmpty, []testCase[mValue]{
		{"non-trivial", mValue{
			1:   empty,
			167: empty,
			4:   empty,
			17:  empty,
		}, nil},
	})
	testCodecFail(t, mapValueEmpty, nil)
}

func TestNegateEmpty(t *testing.T) {
	testCodec(t, negateEmpty, []testCase[emptyStruct]{
		{"neg(empty)", empty, []byte{0xFF}},
	})
	testCodecFail(t, negateEmpty, empty)
}

func TestTerminateEmpty(t *testing.T) {
	testCodec(t, terminateEmpty, []testCase[emptyStruct]{
		{"terminate(empty)", empty, []byte{0x00}},
	})
	testCodecFail(t, terminateEmpty, empty)
}