package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	t.Parallel()
	assert.True(t, lexy.Bytes().RequiresTerminator())
	assert.False(t, lexy.TerminatedBytes().RequiresTerminator())
	testCodec(t, lexy.Bytes(), []testCase[[]byte]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []byte{}, []byte{pNonNil}},
		{"[0]", []byte{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
	})
}

func TestCastBytes(t *testing.T) {
	t.Parallel()
	type myBytes []byte
	codec := lexy.CastBytes[myBytes]()
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[myBytes]{
		{"nil", nil, []byte{pNilFirst}},
		{"empty", []byte{}, []byte{pNonNil}},
		{"[0]", []byte{0}, []byte{pNonNil, 0x00}},
		{"[1, 2, 3]", []byte{1, 2, 3}, []byte{pNonNil, 0x01, 0x02, 0x03}},
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

func TestCastBytesNilsLast(t *testing.T) {
	t.Parallel()
	type myBytes []byte
	codec := lexy.CastBytes[myBytes]()
	testOrdering(t, lexy.NilsLast(codec), []testCase[myBytes]{
		{"empty", []byte{}, nil},
		{"[0]", []byte{0}, nil},
		{"[0, 0, 0]", []byte{0, 0, 0}, nil},
		{"[0, 1]", []byte{0, 1}, nil},
		{"[35]", []byte{35}, nil},
		{"nil", nil, nil},
	})
}
