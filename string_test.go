package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
)

func TestString(t *testing.T) {
	testCodec(t, lexy.String(), []testCase[string]{
		{"empty", "", []byte{}},
		{"a", "a", []byte{'a'}},
		{"xyz", "xyz", []byte{'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
	testCodecFail(t, lexy.String(), "a")
}
