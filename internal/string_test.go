package internal_test

import (
	"testing"
)

func TestString(t *testing.T) {
	codec := stringCodec
	testCodec(t, codec, []testCase[string]{
		{"empty", "", []byte{pEmpty}},
		{"a", "a", []byte{pNonEmpty, 'a'}},
		{"xyz", "xyz", []byte{pNonEmpty, 'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{pNonEmpty, 0xE2, 0x8C, 0x98}},
	})
	testCodecFail(t, codec, "a")
}
