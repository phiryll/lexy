# lexy

Lexicographical Byte Order Encodings

Lexy is a library for encoding/decoding data into byte slices whose
lexicographical unsigned byte ordering is consistent with the data
type's natural ordering. This library is only needed if you need to
order encoded values. Lexy's primary purpose is to make it easier to
use an [ordered key-value
store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store), or an
ordered binary trie.

Lexy will prefix the encoded value with type information, allowing it
to be decoded as `any`. A side effect of this is that encoded values
are ordered by type first and value second. For example, the encodings
of all `int8`s may be less than the encodings of all `int16`s,
regardless of numeric value. Lexy has a default set of type prefix
values which can be overridden to order types differently. The only
way to consistently order the semantic values of different numeric
types is to convert everything to the same exact numeric type before
encoding.

Lexy can encode:

* `bool`  
  `false` sorts before `true`.
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`  
  These are simply encoded in big-endian byte order.
* `int8`, `int16`, `int32` (aka `rune`), `int64`  
  **TODO:** Describe this encoding.
* `float32`, `float64`  
  **TODO:** Describe this encoding.
* `math.big.Int`  
  **TODO:** Describe this encoding.
* `math.big.Float`  
  **TODO:** Describe this encoding.
* `string`  
  A `string` is encoded simply as its bytes. The resulting ordering
  may not reflect the semantic ordering of your use case, because a
  `string` in go is essentially an immutable `[]byte` with no specific
  character encoding, or even text semantics. If your `string`
  contains UTF-8 encoded text, then the encoded ordering will be the
  same as the lexicographical ordering of the corresponding Unicode
  code points. This is not alphabetical, because (for example) the
  code points for "a" and "&#xE9;" will sort after "Z". Collation is
  locale-dependent and Lexy makes no attempt to address this.
* `time.Time`  
  A `time.Time` is encoded as `Time.MarshalText()` of its UTC instant
  followed by its time zone as returned by `Time.Location().String()`.
* slices of supported types  
  Slices are ordered lexicographically.
* maps of supported types  
  Whether map keys are ordered is optional because of the sorting
  overhead. Encoding and decoding will be correct if unordered keys
  are used, but the resulting byte strings will be neither orderable
  nor comparable.
* structs of supported types  
  Lexy cannot access unexported struct fields. Otherwise, structs
  behave similarly to maps with string keys.

Lexy does not currently encode these, but should in the future:

* `nil`  
  Most of the time it might be sufficient to just not encode `nil` at
  all and treat it as an absence of a value. For example, just skip a
  `nil`-valued struct field. However, `nil` maps and slices are not
  the same as empty maps and slices.
* pointer types  
  Pointers are a mandatory use case, especially within slices and
  structs.
* `time.Duration`

Lexy cannot encode these, but you can always write a custom Codec:

* `uint`, `int`, `uintptr`  
  These types have implementation-specific sizes.
* `complex64`
* `complex128`
* `math.big.Rat`  
  While rational numbers are ordered, there is no base in which they
  can be represented at full precision.
* array types
* function types
* interface types
* channel types

## TODO

Determine if there's a way to bridge this with the
`encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces.

Provide an alternate encoder/decoder that omits type information. This
can be used if you know the exact type of what you're decoding. Note
that instances of different types will necessarily be unordered with
respect to each other if you omit type information, and a range query
on a heterogeneous data set could return multiple types.

Provide a mechanism to allow user-specified ordering of struct fields.

Provide some mechanism to handle user-defined types. The user would
need to provide an encoder/decoder for that type, and a type prefix if
using that feature.
