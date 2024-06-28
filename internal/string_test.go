package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestStringCodec(t *testing.T) {
	testCodec[string](t, internal.StringCodec{}, []testCase[string]{
		{"empty", "", []byte{0x03}},
		{"a", "a", []byte{0x04, 'a'}},
		{"xyz", "xyz", []byte{0x04, 'x', 'y', 'z'}},
		{"⌘", "⌘", []byte{0x04, 0xE2, 0x8C, 0x98}},
	})
}
