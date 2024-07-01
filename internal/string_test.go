package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestStringCodec(t *testing.T) {
	codec := internal.StringCodec{}
	testCodec[string](t, codec, []testCase[string]{
		{"empty", "", []byte{zero}},
		{"a", "a", []byte{nonZero, 'a'}},
		{"xyz", "xyz", []byte{nonZero, 'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{nonZero, 0xE2, 0x8C, 0x98}},
	})
	testCodecFail[string](t, codec, "a")
}
