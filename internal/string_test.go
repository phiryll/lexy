package internal

import (
	"testing"
)

func TestStringCodec(t *testing.T) {
	testCodec[string](t, StringCodec{}, []testCase[string]{
		{"empty", "", []byte("")},
		{"a", "a", []byte("a")},
		{"xyz", "xyz", []byte("xyz")},
		{"⌘", "⌘", []byte{0xE2, 0x8C, 0x98}},
	})
}
