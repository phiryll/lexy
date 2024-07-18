# lexy

Lexicographical Byte Order Encodings

Lexy is a library for encoding/decoding strongly-typed values into a binary form whose
lexicographical unsigned byte ordering is consistent with the type's natural ordering.
Users can build their own encoding with a different encoded ordering than the type's natural ordering.
Types defined with a different underlying type will work correctly (`type MyInt int16`, e.g.).
Encoded values of different data types will not have a consistent ordering with each other.
Lexy has no non-test dependencies.

This library is only needed if you need to order encoded values.
Lexy's primary purpose is to make it easier to use things like an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
or an ordered binary trie.

Lexy can encode these types while preserving their natural ordering.

* `bool`  
  `false` is ordered before `true`.
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`
* `int8`, `int16`, `int32` (aka `rune`), `int64`
* `float32`, `float64`
* `*math.big.Int`  
  `nil` is less than all non-`nil` values.
* `*math.big.Float`  
  `nil` is less than all non-`nil` values.
* `string`
* `time.Time`  
  Instances are ordered by UTC time first, timezone offset (at that instant) second.
* `time.Duration`
* pointers  
  `nil` is less than all non-`nil` values.
* slices  
  Slices are ordered lexicographically by their elements.
  For example, `{0, 1} < {0, 1, 100} < {0, 2} < {1}`
* arrays  
  Arrays are ordered lexicographically by their elements.
  For example, `{0, 1, 0} < {0, 1, 100} < {0, 2, 0} < {1, 0, 0}`
  Arrays of different sizes are different types in go.

Lexy can encode these types which either have no natural ordering,
or whose natural ordering cannot be preserved while being encoded at full precision.
The ordering of the encoding is noted.

* maps  
  Maps and their encodings are inherently randomly ordered.
* `complex64`, `complex128`  
  The encoded order is real part first, imaginary part second.
* `*math.big.Rat`  
  `nil` is less than all non-`nil` values.
  The encoded order for non-`nil` values is signed numerator first, positive denominator second.
  There is no way to encode rational numbers with a lexicographical order that isn't lossy.
  The closest you can get is to convert them to (possibly rounded) big.Floats and encode those.

Lexy does not encode these out-of-the-box, but you can always write a custom `Codec`.

* struct types  
  Examples are provided to show how to build your own `Codecs`, including for struct types.
  The inherent limitations of generic types and reflection in go make it impossible
  to do this in a general way without having a parallel, but completely separate, set of non-generic codecs.
  Writing your own `Codec` is a much simpler and safer alternative.
* `uint`, `int`, `uintptr`  
  These types have implementation-specific sizes.
* function types
* interface types
* channel types
