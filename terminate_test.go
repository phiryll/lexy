package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    []byte
		escaped []byte
	}{
		{
			"no special bytes",
			[]byte{2, 3, 5, 4, 7, 6},
			[]byte{2, 3, 5, 4, 7, 6, 0},
		},
		{
			"with special bytes",
			[]byte{0, 1, 2, 3, 1, 4, 0, 5, 6},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 5, 6, 0},
		},
		{
			"empty",
			[]byte{},
			[]byte{0},
		},
		{
			"terminator",
			[]byte{0},
			[]byte{1, 0, 0},
		},
		{
			"escape",
			[]byte{1},
			[]byte{1, 1, 0},
		},
		{
			"trailing terminator",
			[]byte{0, 1, 2, 3, 1, 4, 0},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 0},
		},
		{
			"trailing escape",
			[]byte{0, 1, 2, 3, 1, 4, 1},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1, 0},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := lexy.TestingDoEscape(tt.data)
			assert.Equal(t, tt.escaped, buf, "escaped bytes")
		})
	}
}

//nolint:funlen
func TestUnescape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		data      []byte
		unescaped []byte // nil if a panic is expected
	}{
		{
			"no special bytes",
			[]byte{2, 3, 5, 4, 7, 6},
			nil,
		},
		{
			"with special bytes",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 5, 6},
			nil,
		},
		{
			"empty",
			[]byte{},
			nil,
		},
		{
			"terminator",
			[]byte{1, 0},
			nil,
		},
		{
			"escape",
			[]byte{1, 1},
			nil,
		},
		{
			"trailing escaped terminator",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0},
			nil,
		},
		{
			"trailing escaped escape",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1},
			nil,
		},
		{
			"trailing unescaped terminator",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 0},
			[]byte{0, 1, 2, 3, 1, 4},
		},
		{
			"non-trailing unescaped terminator",
			[]byte{2, 3, 4, 1, 0, 5, 6, 0, 7, 8, 9},
			[]byte{2, 3, 4, 0, 5, 6},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.unescaped == nil {
				assert.Panics(t, func() {
					lexy.TestingDoUnescape(tt.data)
				})
			} else {
				buf, _, _ := lexy.TestingDoUnescape(tt.data)
				assert.Equal(t, tt.unescaped, buf)
			}
		})
	}
}

func TestUnescapeMultiple(t *testing.T) {
	t.Parallel()
	data := []byte{2, 3, 1, 0, 5, 0, 7, 8, 9, 0, 10, 11, 12, 0}
	n := 0
	for _, expected := range [][]byte{
		{2, 3, 0, 5},
		{7, 8, 9},
		{10, 11, 12},
	} {
		var got []byte
		got, data, _ = lexy.TestingDoUnescape(data)
		assert.Equal(t, expected, got, "unescaped bytes")
	}
	assert.Len(t, data, n)
}
