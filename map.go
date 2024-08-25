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
	done, buf := c.prefix.Append(buf, value == nil)
	if done {
		return buf
	}
	for k, v := range value {
		buf = c.keyCodec.Append(buf, k)
		buf = c.valueCodec.Append(buf, v)
	}
	return buf
}

func (c mapCodec[K, V]) Put(buf []byte, value map[K]V) []byte {
	done, buf := c.prefix.Put(buf, value == nil)
	if done {
		return buf
	}
	for k, v := range value {
		buf = c.keyCodec.Put(buf, k)
		buf = c.valueCodec.Put(buf, v)
	}
	return buf
}

func (c mapCodec[K, V]) Get(buf []byte) (map[K]V, []byte) {
	done, buf := c.prefix.Get(buf)
	if done {
		return nil, buf
	}
	m := map[K]V{}
	var key K
	var value V
	for {
		if len(buf) == 0 {
			return m, buf
		}
		key, buf = c.keyCodec.Get(buf)
		value, buf = c.valueCodec.Get(buf)
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
