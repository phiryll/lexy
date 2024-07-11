package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestString(t *testing.T) {
	codec := internal.StringCodec
	testCodec(t, codec, []testCase[string]{
		{"empty", "", []byte{empty}},
		{"a", "a", []byte{nonEmpty, 'a'}},
		{"xyz", "xyz", []byte{nonEmpty, 'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{nonEmpty, 0xE2, 0x8C, 0x98}},
	})
	testCodecFail(t, codec, "a")
}
