package internal_test

import (
	"testing"
)

func TestString(t *testing.T) {
	codec := stringCodec
	testCodec(t, codec, []testCase[string]{
		{"empty", "", []byte{}},
		{"a", "a", []byte{'a'}},
		{"xyz", "xyz", []byte{'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
	testCodecFail(t, codec, "a")
}
