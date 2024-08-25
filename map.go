package lexy

// mapCodec is the unordered Codec for maps.
// A map is encoded as:
//
// - if nil, prefixNilFirst/Last
// - if non-nil, prefixNonNil, encoded key, encoded value, encoded key, ...
//
// Encoded keys and values are escaped and termninated if their respective Codecs require it.
type mapCodec[K comparable, V any] struct {
	keyCodec   Codec[K]
	valueCodec Codec[V]
	prefix     Prefix
}

func (c mapCodec[K, V]) Append(buf []byte, value map[K]V) []byte {
	done, newBuf := c.prefix.Append(buf, value == nil)
	if done {
		return newBuf
	}
	for k, v := range value {
		newBuf = c.keyCodec.Append(newBuf, k)
		newBuf = c.valueCodec.Append(newBuf, v)
	}
	return newBuf
}

func (c mapCodec[K, V]) Put(buf []byte, value map[K]V) int {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return 1
	}
	n := 0
	for k, v := range value {
		n += c.keyCodec.Put(buf[n:], k)
		n += c.valueCodec.Put(buf[n:], v)
	}
	return 1 + n
}

func (c mapCodec[K, V]) Get(buf []byte) (map[K]V, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	m := map[K]V{}
	for {
		if len(buf) == 0 {
			return m, buf
		}
		key, newBuf := c.keyCodec.Get(buf)
		value, newBuf := c.valueCodec.Get(newBuf)
		buf = newBuf
		m[key] = value
	}
}

func (mapCodec[K, V]) RequiresTerminator() bool {
	return true
}

//lint:ignore U1000 this is actually used
func (c mapCodec[K, V]) nilsLast() Codec[map[K]V] {
	return mapCodec[K, V]{c.keyCodec, c.valueCodec, PrefixNilsLast}
}
