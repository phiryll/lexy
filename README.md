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

* `nil`
* `bool`  
  `false` sorts before `true`.
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`
* `int8`, `int16`, `int32` (aka `rune`), `int64`
* `float32`, `float64`
* `math.big.Int`
* `math.big.Float`
* `string`
* `time.Time`  
  Time instances are sorted by UTC time first, timezone second.
* slices of supported types
* maps of supported types  
  Whether map keys are ordered is optional because of the sorting overhead.
  Encoding and decoding will be correct if unordered keys are used,
  but the results will be randomly ordered.
* structs of supported types  
  Lexy cannot access unexported struct fields.
  Otherwise, structs behave similarly to maps with string keys.

Lexy does not currently encode these, but should in the future:

* `time.Duration`
* pointer types

Lexy cannot encode these, but you can always write a custom Codec:

* `uint`, `int`, `uintptr`  
  These types have implementation-specific sizes.
* `complex64`
* `complex128`
* `math.big.Rat`
* array types
* function types
* interface types
* channel types
