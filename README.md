# lexy

[![Build Status](https://github.com/phiryll/lexy/actions/workflows/tests.yaml/badge.svg?branch=main)](https://github.com/phiryll/lexy/actions/workflows/tests.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/phiryll/lexy)](https://goreportcard.com/report/github.com/phiryll/lexy)
[![Go Reference](https://pkg.go.dev/badge/github.com/phiryll/lexy)](https://pkg.go.dev/github.com/phiryll/lexy)

Lexicographical Byte Order Encodings

Lexy is a library for encoding strongly-typed values (using generics) into a binary form whose
lexicographical unsigned byte ordering is consistent with the type's natural ordering.
Custom encodings can be created with a different ordering than the type's natural ordering.
Lexy has no non-test dependencies.

It may be more efficient to use another encoding if lexicographical unsigned byte ordering is not needed.
Lexy's primary purpose is to make it easier to use an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
an ordered binary trie, or similar.

The primary interface in lexy is `Codec`, with this definition (details in go docs):

```go
type Codec[T any] interface {
    // Read reads from r and decodes a value of type T.
    Read(r io.Reader) (T, error)

    // Write encodes value and writes the encoded bytes to w.
    Write(w io.Writer, value T) error

    // RequiresTerminator must return true if Read may not know
    // when to stop reading the data encoded by Write,
    // or if Write could encode zero bytes for some value.
    // This is the case for unbounded types like strings, slices,
    // and maps, as well as empty struct types.
    RequiresTerminator() bool
}
```

A typical use might look something like this:

```go
type Word string
type Key []Word
type Value struct {
  // ...
}

// keyCodec is safe for concurrent use.
var keyCodec = lexy.SliceOf[Key](lexy.String[Word]())

// lexy could be used here, but it's overkill if ordered Values aren't needed.
func EncodeValue(v *Value) ([]byte, error) { /* ... */ }
func DecodeValue(b []byte) (*Value, error) { /* ... */ }

type KeyValueDB struct {
    providerDB *provider.DB
    // ...
}

func (db *KeyValueDB) Put(key Key, value *Value) error {
    var buf bytes.Buffer
    if err := keyCodec.Write(&buf, key); err != nil {
        return err
    }
    valueBytes, err := EncodeValue(value)
    if err != nil {
        return err
    }
    // buf.Bytes() can be nil if no bytes were written by keyCodec.
    return db.providerDB.Put(buf.Bytes(), valueBytes)
}

func (db *KeyValueDB) Get(key Key) (*Value, error) {
    var buf bytes.Buffer
    if err := keyCodec.Write(&buf, key); err != nil {
        return nil, err
    }
    // buf.Bytes() can be nil if no bytes were written by keyCodec.
    valueBytes, err := db.providerDB.Get(buf.Bytes())
    if err != nil {
        return nil, err
    }
    return DecodeValue(valueBytes)
}
```

All `Codecs` provided by lexy are safe for concurrent use if their delegate `Codecs` (if any) are.

Lexy provides `Codecs` for these types that preserve their natural ordering.

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
  `nil` can be less than or greater than all non-`nil` values.
* `*math.big.Float`  
  `nil` can be less than or greater than all non-`nil` values.
* `string`
* `time.Time`  
  Instances are ordered by UTC time first, timezone offset (at that instant) second.
* `time.Duration`
* pointers  
  `nil` can be less than or greater than all non-`nil` values.
* slices  
  `nil` can be less than or greater than all non-`nil` values.
  Slices are ordered lexicographically by their elements.
  For example,  
  `{0, 1} < {0, 1, 100} < {0, 2} < {1}`
* `[]byte`  
  `nil` can be less than or greater than all non-`nil` values.
  This `Codec` is optimized for byte slices, and is more efficient than a slice `Codec` would be.
  It differs from the `string` `Codec` in that a `[]byte` can be `nil`.

Lexy provides `Codecs` for these types which either have no natural ordering,
or whose natural ordering cannot be preserved while being encoded at full precision.

* maps  
  `nil` can be less than or greater than all non-`nil` maps.
  Empty maps are always less than non-empty maps.
  Non-empty maps are randomly ordered.
* `complex64`, `complex128`  
  The encoded order is real part first, imaginary part second.
* `*math.big.Rat`  
  `nil` can be less than or greater than all non-`nil` values.
  The encoded order for non-`nil` values is signed numerator first, positive denominator second.
  There is no way to finitely encode rational numbers with a lexicographical order that isn't lossy.
  A lossy approximation can be made by converting to (possibly rounded) `big.Floats` and encoding those.

Lexy provides a `Codec` for types with no value except the zero value,
useful as value types for maps used as sets.

Lexy does not does not provide `Codecs` for these types, but a custom `Codec` is easy to create.
See the provided examples for how to create custom `Codecs`.

* structs, pointers to structs  
  The inherent limitations of generic types and reflection in go make it impossible
  to do this in a general way without having a parallel, but completely separate, set of non-generic codecs.
  Writing a strongly-typed custom `Codec` is a much simpler and safer alternative,
  and also prevents silently changing an encoding when the data type it encodes is changed.
* arrays, pointers to arrays  
  While it is possible to create a general `Codec` for array types,
  the generics are very messy and it requires using reflection extensively.
  As is the case for structs, writing a strongly-typed custom `Codec` is a better option.
* `uintptr`  
  This type has an implementation-specific size.
* functions
* interfaces
* channels

The provided `Codecs` do not encode the types of encoded data, users must know what is being decoded.
A custom `Codec` that handles multiple types could be created,
but it would require a concrete wrapper type to conform to the `Codec[T]` interface.

Types defined with a different underlying type will work correctly if the `Codec` is defined appropriately.
For example, values of type `type MyInt int16` can be used with a `Codec` created by `lexy.Int16[MyInt]()`.

Encoded values of different data types will not have a consistent ordering with respect to each other.
For example, the encoded value of `int32(1)` is greater than the encoded value of `uint32(2)`.
