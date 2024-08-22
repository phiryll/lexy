package lexy

import (
	"io"
)

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its referent.
// A pointer is encoded as:
//   - if nil, prefixNilFirst/Last
//   - if non-nil, prefixNonNil followed by its encoded referent
type pointerCodec[E any] struct {
	elemCodec Codec[E]
	prefix    Prefix
}

func (c pointerCodec[E]) Append(buf []byte, value *E) []byte {
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	return c.elemCodec.Append(newBuf, *value)
}

func (c pointerCodec[E]) Put(buf []byte, value *E) int {
	if c.prefix.Put(buf, value == nil) {
		return 1
	}
	n := 1
	return n + c.elemCodec.Put(buf[n:], *value)
}

func (c pointerCodec[E]) Get(buf []byte) (*E, int) {
	if len(buf) == 0 {
		return nil, -1
	}
	if c.prefix.Get(buf) {
		return nil, 1
	}
	n := 1
	value, count := c.elemCodec.Get(buf[n:])
	n += count
	if count < 0 {
		panic(io.ErrUnexpectedEOF)
	}
	return &value, n
}

func (c pointerCodec[E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c pointerCodec[E]) nilsLast() Codec[*E] {
	return pointerCodec[E]{c.elemCodec, PrefixNilsLast}
}
