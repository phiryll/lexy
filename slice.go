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
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	for _, elem := range value {
		buf = c.elemCodec.Append(buf, elem)
	}
	return buf
}

func (c sliceCodec[E]) Put(buf []byte, value []E) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	for _, elem := range value {
		buf = c.elemCodec.Put(buf, elem)
	}
	return buf
}

func (c sliceCodec[E]) Get(buf []byte) ([]E, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	values := []E{}
	var value E
	for {
		if len(buf) == 0 {
			return values, buf
		}
		value, buf = c.elemCodec.Get(buf)
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
