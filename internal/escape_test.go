package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		escaped []byte
	}{
		{"no special bytes",
			[]byte{2, 3, 5, 4, 7, 6},
			[]byte{2, 3, 5, 4, 7, 6}},
		{"with special bytes",
			[]byte{0, 1, 2, 3, 1, 4, 0, 5, 6},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 5, 6}},
		{"empty",
			[]byte{},
			[]byte(nil)}, // buf.Bytes() returns nil if not written to
		{"delimiter",
			[]byte{0},
			[]byte{1, 0}},
		{"escape",
			[]byte{1},
			[]byte{1, 1}},
		{"trailing delimiter",
			[]byte{0, 1, 2, 3, 1, 4, 0},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0}},
		{"trailing escape",
			[]byte{0, 1, 2, 3, 1, 4, 1},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			count, err := internal.Escape(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, len(tt.data), count, "bytes read from input")
			assert.Equal(t, tt.escaped, buf.Bytes(), "escaped bytes")
		})
	}
}

func TestEscapeFail(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		escaped []byte
		count   int
		wantErr bool
	}{
		{"no special bytes, at limit",
			[]byte{2, 3, 5, 4, 7},
			[]byte{2, 3, 5, 4, 7},
			5,
			false},
		{"no special bytes, over limit",
			[]byte{2, 3, 5, 4, 7, 6},
			[]byte{2, 3, 5, 4, 7},
			5,
			true},
		{"with special bytes, at limit",
			[]byte{0, 1, 2},
			[]byte{1, 0, 1, 1, 2},
			3,
			false},
		{"with special bytes, over limit",
			[]byte{0, 1, 2, 3, 1, 4, 0, 5, 6},
			[]byte{1, 0, 1, 1, 2},
			3,
			true},
		{"special at limit",
			[]byte{2, 3, 4, 0},
			[]byte{2, 3, 4, 1, 0},
			4,
			false},
		{"escaped crosses limit",
			[]byte{2, 3, 4, 5, 0},
			[]byte{2, 3, 4, 5, 1},
			4,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := boundedWriter{limit: 5}
			count, err := internal.Escape(&w, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, count, "bytes read from input")
			assert.Equal(t, tt.escaped, w.data, "escaped bytes")
		})
	}
}

func TestUnescape(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		unescaped []byte
		atEof     bool
	}{
		{"no special bytes",
			[]byte{2, 3, 5, 4, 7, 6},
			[]byte{2, 3, 5, 4, 7, 6},
			true},
		{"with special bytes",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 5, 6},
			[]byte{0, 1, 2, 3, 1, 4, 0, 5, 6},
			true},
		{"empty",
			[]byte{},
			[]byte(nil), // buf.Bytes() returns nil if not written to
			true},
		{"delimiter",
			[]byte{1, 0},
			[]byte{0},
			true},
		{"escape",
			[]byte{1, 1},
			[]byte{1},
			true},
		{"trailing escaped delimiter",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0},
			[]byte{0, 1, 2, 3, 1, 4, 0},
			true},
		{"trailing escaped escape",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1},
			[]byte{0, 1, 2, 3, 1, 4, 1},
			true},
		{"trailing unescaped delimiter",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 0},
			[]byte{0, 1, 2, 3, 1, 4},
			false},
		// This case is malformed, but testing expected behavior (white-box testing here).
		{"trailing unescaped escape",
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1},
			[]byte{0, 1, 2, 3, 1, 4},
			true},
		{"non-trailing unescaped delimiter",
			[]byte{2, 3, 4, 1, 0, 5, 6, 0, 7, 8, 9},
			[]byte{2, 3, 4, 0, 5, 6},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			got, err := internal.Unescape(r)
			if tt.atEof {
				assert.ErrorIs(t, err, io.EOF)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.unescaped, got, "unescaped bytes")
		})
	}
}

func TestUnescapeMultiple(t *testing.T) {
	data := []byte{2, 3, 1, 0, 5, 0, 7, 8, 9, 0, 10, 11, 12, 0}
	r := bytes.NewReader(data)

	for _, expected := range [][]byte{
		{2, 3, 0, 5},
		{7, 8, 9},
		{10, 11, 12},
	} {
		got, err := internal.Unescape(r)
		require.NoError(t, err)
		assert.Equal(t, expected, got, "unescaped bytes")
	}
	got, err := internal.Unescape(r)
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, []byte(nil), got, "exhausted")
}
