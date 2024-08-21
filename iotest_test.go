package lexy_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helpers related to io.Reader/Writer behavior,
// things not in the std lib iotest package.
// There are no top-level tests of non-test functionality here.

var (
	errRead  = errors.New("failed to read")
	errWrite = errors.New("failed to write")
)

// The same as iotest.TruncateWriter, except it's not silent.
type boundedWriter struct {
	w         io.Writer
	remaining int
}

func (t *boundedWriter) Write(p []byte) (int, error) {
	if t.remaining <= 0 {
		return 0, errWrite
	}
	// real write
	n := len(p)
	if n > t.remaining {
		n = t.remaining
	}
	n, err := t.w.Write(p[0:n])
	t.remaining -= n
	if err != nil {
		// some other error, return it
		return n, err
	}
	if n < len(p) && t.remaining == 0 {
		return n, errWrite
	}
	return n, nil
}

func TestBoundedWriter(t *testing.T) {
	// no failure if below the limit
	buf := bytes.NewBuffer([]byte{})
	w := boundedWriter{buf, 10}
	n, err := w.Write(make([]byte, 10))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)

	// failure if read one more byte
	n, err = w.Write([]byte{0})
	assert.Error(t, err)
	assert.Equal(t, 0, n)

	// reset, failure if read one over limit
	buf.Reset()
	w = boundedWriter{buf, 10}
	n, err = w.Write(make([]byte, 11))
	assert.Error(t, err)
	assert.Equal(t, 10, n)
}
