package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	t.Parallel()
	assert.True(t, lexy.String().RequiresTerminator())
	assert.False(t, lexy.TerminatedString().RequiresTerminator())
	testCodec(t, lexy.String(), []testCase[string]{
		{"empty", "", []byte{}},
		{"a", "a", []byte{'a'}},
		{"xyz", "xyz", []byte{'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
}

func TestCastString(t *testing.T) {
	t.Parallel()
	type myString string
	codec := lexy.CastString[myString]()
	assert.True(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[myString]{
		{"empty", "", []byte{}},
		{"a", "a", []byte{'a'}},
		{"xyz", "xyz", []byte{'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
}
