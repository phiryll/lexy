package lexy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests that don't fit better somewhere else.

// Tests that the bounds check idiom being used by lexy doesn't get optimized out by the Go compiler.
func TestBoundsCheck(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		buf := make([]byte, 10)
		_ = buf[10]
	})
}

// Demonstrating a gotcha when working with slices.
// A slice can be resliced up to it's capacity, regardless of its length.
func TestIndexingBeyondSliceLength(t *testing.T) {
	t.Parallel()
	buf := []byte{0, 1, 2, 3}
	empty := buf[:0]

	assert.Panics(t, func() {
		_ = empty[0]
	})
	assert.Equal(t, byte(3), (empty[:4])[3])

	assert.Panics(t, func() {
		_ = empty[1:]
	})
	assert.Equal(t, byte(3), (empty[1:4])[2])
}
