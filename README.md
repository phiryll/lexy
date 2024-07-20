# lexy

Lexicographical Byte Order Encodings

Lexy is a library for encoding strongly-typed values (using generics) into a binary form whose
lexicographical unsigned byte ordering is consistent with the type's natural ordering.
Custom encodings can be created with a different ordering than the type's natural ordering.
Lexy has no non-test dependencies.

It may be more efficient to use another encoding if lexicographical unsigned byte ordering is not needed.
Lexy's primary purpose is to make it easier to use things like an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
or an ordered binary trie.

Lexy does not encode the types of encoded data, users must know what is being decoded.
Types defined with a different underlying type will work correctly if the `Codec` is defined appropriately.
For example, values of type `type MyInt int16` can be used with a `Codec` created by `lexy.Int[MyInt]()`.
Encoded values of different data types will not have a consistent ordering with respect to each other.
For example, the encoded value of `int32(1)` is greater than the encoded value of `uint32(2)`.
A custom `Codec` that handles multiple types could be created,
but it would require a concrete wrapper type to conform to the `Codec[T]` interface.

Lexy can encode these types while preserving their natural ordering.

* `bool`  
  `false` is ordered before `true`.
* `uint`  
  Instances are encoded using the `uint64` Codec.
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`
* `int`  
  Instances are encoded using the `int64` Codec.
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
  For example,  
  `{0, 1} < {0, 1, 100} < {0, 2} < {1}`
* `[]byte`  
  This is a `Codec` optimized for byte slices, more efficient than a slice `Codec` would be.
  It differs from the `string` `Codec` in that a `[]byte` can be `nil`.
* arrays  
  Arrays are ordered lexicographically by their elements.
  For example,  
  `{0, 1, 0} < {0, 1, 100} < {0, 2, 0} < {1, 0, 0}`  
  Arrays of different sizes are different types in go, and will require different `Codecs`.
  The `Codec` created by `lexy.ArrayOf` makes heavy use of reflection, and should be avoided if possible.
  See the provided examples for how to create custom `Codecs`.

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
  There is no way to finitely encode rational numbers with a lexicographical order that isn't lossy.
  A lossy approximation can be made by converting to (possibly rounded) `big.Floats` and encoding those.

Lexy does not encode these out-of-the-box, but a custom `Codec` can always be created.

* struct types  
  Examples are provided to show how to build custom `Codecs`, including for struct types.
  The inherent limitations of generic types and reflection in go make it impossible
  to do this in a general way without having a parallel, but completely separate, set of non-generic codecs.
  Writing a custom `Codec` is a much simpler and safer alternative.
* `uintptr`  
  This type has an implementation-specific size.
* function types
* interface types
* channel types
