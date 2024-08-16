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
