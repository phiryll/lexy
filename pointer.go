package lexy

// pointerCodec is the Codec for pointers, using elemCodec to encode and decode its referent.
// A pointer is encoded as:
//   - if nil, prefixNilFirst/Last
//   - if non-nil, prefixNonNil followed by its encoded referent
type pointerCodec[E any] struct {
	elemCodec Codec[E]
	prefix    Prefix
}

func (c pointerCodec[E]) Append(buf []byte, value *E) []byte {
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	return c.elemCodec.Append(buf, *value)
}

func (c pointerCodec[E]) Put(buf []byte, value *E) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	return c.elemCodec.Put(buf, *value)
}

func (c pointerCodec[E]) Get(buf []byte) (*E, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	value, buf := c.elemCodec.Get(buf)
	return &value, buf
}

func (c pointerCodec[E]) RequiresTerminator() bool {
	return c.elemCodec.RequiresTerminator()
}

//lint:ignore U1000 this is actually used
func (c pointerCodec[E]) nilsLast() Codec[*E] {
	return pointerCodec[E]{c.elemCodec, PrefixNilsLast}
}
