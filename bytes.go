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
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	return append(newBuf, value...)
}

func (c bytesCodec) Put(buf, value []byte) int {
	return mustCopy(buf, c.Append(nil, value))
}

func (c bytesCodec) Get(buf []byte) ([]byte, int) {
	if len(buf) == 0 {
		return nil, -1
	}
	if c.prefix.Get(buf) {
		return nil, 1
	}
	return append([]byte{}, buf[1:]...), len(buf)
}

func (bytesCodec) RequiresTerminator() bool {
	return true
}

//lint:ignore U1000 this is actually used
func (bytesCodec) nilsLast() Codec[[]byte] {
	return bytesCodec{PrefixNilsLast}
}
