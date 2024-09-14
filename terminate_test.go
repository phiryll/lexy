package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

// A []byte codec that does nothing, encoded == decoded,
// purely for testing terminatorCodec.
type nopCodec struct{}

var nop lexy.Codec[[]byte] = nopCodec{}

func (nopCodec) Append(buf, value []byte) []byte {
	return append(buf, value...)
}

func (nopCodec) Put(buf, value []byte) []byte {
	if len(value) == 0 {
		return buf
	}
	_ = buf[len(value)-1]
	return buf[copy(buf, value):]
}

func (nopCodec) Get(buf []byte) ([]byte, []byte) {
	return append([]byte{}, buf...), buf[len(buf):]
}

func (nopCodec) RequiresTerminator() bool {
	return true
}

func TestTerminator(t *testing.T) {
	t.Parallel()
	codec := lexy.Terminate(nop)
	assert.False(t, codec.RequiresTerminator())
	testCodec(t, codec, []testCase[[]byte]{
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
			"trailing terminator",
			[]byte{0, 1, 2, 3, 1, 4, 0},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 0},
		},
		{
			"trailing escape",
			[]byte{0, 1, 2, 3, 1, 4, 1},
			[]byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1, 0},
		},
	})
}

func TestUnescapePanic(t *testing.T) {
	t.Parallel()
	codec := lexy.Terminate(nop)
	for _, tt := range []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"terminator", []byte{1, 0}},
		{"escape", []byte{1, 1}},
		{"no special bytes", []byte{2, 3, 5, 4, 7, 6}},
		{"with special bytes", []byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0, 5, 6}},
		{"trailing escaped terminator", []byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 0}},
		{"trailing escaped escape", []byte{1, 0, 1, 1, 2, 3, 1, 1, 4, 1, 1}},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Panics(t, func() {
				codec.Get(tt.data)
			})
		})
	}
}

func TestUnescapeMultiple(t *testing.T) {
	t.Parallel()
	codec := lexy.Terminate(nop)
	data := []byte{2, 3, 1, 0, 5, 0, 7, 8, 9, 0, 10, 11, 12, 0}
	n := 0
	for _, expected := range [][]byte{
		{2, 3, 0, 5},
		{7, 8, 9},
		{10, 11, 12},
	} {
		var got []byte
		got, data = codec.Get(data)
		assert.Equal(t, expected, got, "unescaped bytes")
	}
	assert.Len(t, data, n)
}
