package lexy

// sliceCodec is the Codec for slices, using elemCodec to encode and decode its elements.
// A slice is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil followed by its encoded elements
//
// Encoded elements are escaped and termninated if elemCodec requires it.
type sliceCodec[E any] struct {
	elemCodec Codec[E]
	prefix    Prefix
}

func (c sliceCodec[E]) Append(buf []byte, value []E) []byte {
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	for _, elem := range value {
		newBuf = c.elemCodec.Append(newBuf, elem)
	}
	return newBuf
}

func (c sliceCodec[E]) Put(buf []byte, value []E) int {
	if c.prefix.Put(buf, value == nil) {
		return 1
	}
	n := 1
	for _, elem := range value {
		n += c.elemCodec.Put(buf[n:], elem)
	}
	return n
}

func (c sliceCodec[E]) Get(buf []byte) ([]E, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	values := []E{}
	for {
		if len(buf) == 0 {
			return values, buf
		}
		value, newBuf := c.elemCodec.Get(buf)
		buf = newBuf
		values = append(values, value)
	}
}

func (sliceCodec[E]) RequiresTerminator() bool {
	return true
}

//lint:ignore U1000 this is actually used
func (c sliceCodec[E]) nilsLast() Codec[[]E] {
	return sliceCodec[E]{c.elemCodec, PrefixNilsLast}
}
