package lexy_test

import (
	"testing"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
)

func TestPrefixNilsFirst(t *testing.T) {
	t.Parallel()
	testPrefix(t, lexy.PrefixNilsFirst, pNilFirst)
}

func TestPrefixNilsLast(t *testing.T) {
	t.Parallel()
	testPrefix(t, lexy.PrefixNilsLast, pNilLast)
}

func testPrefix(t *testing.T, prefix lexy.Prefix, goodPrefix byte) {
	badPrefix := pNilLast
	if goodPrefix == pNilLast {
		badPrefix = pNilFirst
	}

	done, buf := prefix.Append(nil, true)
	assert.True(t, done)
	assert.Equal(t, []byte{goodPrefix}, buf)

	done, buf = prefix.Append(nil, false)
	assert.False(t, done)
	assert.Equal(t, []byte{pNonNil}, buf)

	buf = []byte{0, 1, 2, 3}
	done, newBuf := prefix.Put(buf, true)
	assert.True(t, done)
	assert.Equal(t, []byte{goodPrefix, 1, 2, 3}, buf)
	assert.Equal(t, buf[1:], newBuf)

	done, newBuf = prefix.Put(buf, false)
	assert.False(t, done)
	assert.Equal(t, []byte{pNonNil, 1, 2, 3}, buf)
	assert.Equal(t, buf[1:], newBuf)

	assert.Panics(t, func() {
		prefix.Put(nil, true)
	})
	assert.Panics(t, func() {
		prefix.Put([]byte{}, true)
	})

	buf = []byte{pNonNil, 4, 5, 6}
	done, newBuf = prefix.Get(buf)
	assert.False(t, done)
	assert.Equal(t, buf[1:], newBuf)

	buf = []byte{goodPrefix, 4, 5, 6}
	done, newBuf = prefix.Get(buf)
	assert.True(t, done)
	assert.Equal(t, buf[1:], newBuf)

	assert.Panics(t, func() {
		prefix.Get(nil)
	})
	assert.Panics(t, func() {
		prefix.Get([]byte{})
	})
	assert.Panics(t, func() {
		prefix.Get([]byte{badPrefix, 4, 5, 6})
	})
	msg := getPanicMessage(func() { prefix.Get([]byte{0x37}) })
	assert.Contains(t, msg, "0x37")
}
