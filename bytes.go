package lexy

// bytesCodec is the Codec for []byte.
//
// Get will fully consume its argument buffer if the value is not nil.
// []byte is slightly different than string because it can be nil.
// This is more efficient than sliceCodec would be.
type bytesCodec struct {
	prefix Prefix
}

func (c bytesCodec) Append(buf, value []byte) []byte {
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	return append(buf, value...)
}

func (c bytesCodec) Put(buf, value []byte) []byte {
	return copyAll(buf, c.Append(nil, value))
}

func (c bytesCodec) Get(buf []byte) ([]byte, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	return append([]byte{}, buf...), buf[len(buf):]
}

func (bytesCodec) RequiresTerminator() bool {
	return true
}

//lint:ignore U1000 this is actually used
func (bytesCodec) nilsLast() Codec[[]byte] {
	return bytesCodec{PrefixNilsLast}
}
