package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestStringCodec(t *testing.T) {
	testCodec[string](t, internal.StringCodec{}, []testCase[string]{
		{"empty", "", []byte{zero}},
		{"a", "a", []byte{nonZero, 'a'}},
		{"xyz", "xyz", []byte{nonZero, 'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{nonZero, 0xE2, 0x8C, 0x98}},
	})
}
