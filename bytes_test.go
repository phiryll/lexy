package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
)

func TestBytes(t *testing.T) {
	t.Parallel()
	testCodec(t, lexy.Bytes(), []testCase[[]byte]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []byte{}, []byte{pNonNil}},
		{"[0]", []byte{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
}

func TestBytesUnderlyingType(t *testing.T) {
	t.Parallel()
	type header []byte
	codec := lexy.NilsLast(lexy.CastBytes[header]())
	testCodec(t, codec, []testCase[header]{
		{"nil", header(nil), []byte{pNilLast}},
		{"empty", header{}, []byte{pNonNil}},
		{"[0]", header{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", header{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
}

func TestBytesNilsLast(t *testing.T) {
	t.Parallel()
	testOrdering(t, lexy.NilsLast(lexy.Bytes()), []testCase[[]byte]{
		{"empty", []byte{}, nil},
		{"[0]", []byte{0}, nil},
		{"[0, 0, 0]", []byte{0, 0, 0}, nil},
		{"[0, 1]", []byte{0, 1}, nil},
		{"[35]", []byte{35}, nil},
		{"nil", nil, nil},
	})
}
