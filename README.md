# lexy

Lexicographical Byte Order Encodings

Lexy is a library for encoding/decoding data into a binary form whose
lexicographical unsigned byte ordering is consistent with the data type's natural ordering.
Encoded values from different data types will not have a consistent ordering.
Lexy intentionally has no non-test dependencies.

This library is only needed if you need to order encoded values.
Lexy's primary purpose is to make it easier to use an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
or an ordered binary trie.

Lexy can encode:

* `bool`  
  `false` sorts before `true`.
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`
* `int8`, `int16`, `int32` (aka `rune`), `int64`
* `float32`, `float64`
* `*math.big.Int`
* `*math.big.Float`
* `string`
* `time.Time`  
  Time instances are sorted by UTC time first, timezone offset (at that instant) second.
* `time.Duration`
* pointers to supported types
* slices of supported types
* maps of supported types  
  Maps are inherently unordered, but are included because they are such a basic type.

Lexy provides functionality to help clients to write their own Codecs for struct types.
The inherent limitations of generic types and reflection in go make it impossible
to do this in a general way without having a parallel, but completely separate, set of non-generic Codecs.

Lexy does not encode these, but you can always write a custom Codec:

* `uint`, `int`, `uintptr`  
  These types have implementation-specific sizes.
* `complex64`, `complex128`  
  Complex types have no commonly understood ordering.
* `math.big.Rat`  
  There is no good way to encode rational numbers with a lexicographical order that isn't lossy.
  The closest you can get is to convert them to (possibly rounded) big.Floats and encode those.
* array types
* function types
* interface types
* channel types
