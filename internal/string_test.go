package internal_test

import (
	"testing"

	"github.com/phiryll/lexy/internal"
)

func TestStringCodec(t *testing.T) {
	testCodec[string](t, internal.StringCodec{}, []testCase[string]{
		{"empty", "", []byte(nil)},
		{"a", "a", []byte("a")},
		{"xyz", "xyz", []byte("xyz")},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
}
