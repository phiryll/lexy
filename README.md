# lexy

[![Build Status](https://github.com/phiryll/lexy/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/phiryll/lexy/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/phiryll/lexy)](https://goreportcard.com/report/github.com/phiryll/lexy)
[![Go Reference](https://pkg.go.dev/badge/github.com/phiryll/lexy)](https://pkg.go.dev/github.com/phiryll/lexy)

Lexicographical Byte Order Encodings

Lexy is a library for order-preserving lexicographical binary encodings.
Most common Go types and user-defined types are supported,
and it allows for encodings ordered differently than a type's natural ordering.
Lexy uses generics and requires Go 1.19 to use. It has been tested through Go 1.22.
Lexy has no non-test dependencies.

It may be more efficient to use another encoding if lexicographical unsigned byte ordering is not needed.
Lexy's primary purpose is to make it easier to use an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
an ordered binary trie, or similar.

The primary interface in lexy is `Codec`, with this definition (details in Go docs):

```go
type Codec[T any] interface {
    // Append encodes value and appends the encoded bytes to buf,
    // returning the updated buffer.
    Append(buf []byte, value T) []byte

    // Put encodes value into buf,
    // returning buf following what was written.
    Put(buf []byte, value T) []byte

    // Get decodes a value of type T from buf,
    // returning the value and buf following the encoded value.
    Get(buf []byte) (T, []byte)

    // RequiresTerminator returns whether encoded values require
    // a terminator and escaping if more data is written following
    // the encoded value. This is the case for most unbounded types
    // like strings and slices, as well as types whose encodings
    // can be zero bytes.
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
// The terser functions lexy.SliceOf and lexy.String can be used
// if the types involved are the same as their underlying types,
// string and []string in this case. That would look like this:
//
//   var keyCodec = lexy.SliceOf(lexy.String())
//
var keyCodec = lexy.CastSliceOf[Key](lexy.CastString[Word]())

// lexy could be used here,
// but may be overkill if ordered Values aren't needed.
func EncodeValue(v *Value) ([]byte, error) { /* ... */ }
func DecodeValue(b []byte) (*Value, error) { /* ... */ }

type KeyValueDB struct {
    providerDB *provider.DB
    // ...
}

func (db *KeyValueDB) Put(key Key, value *Value) error {
    // If keyCodec could encode zero bytes, this might be preferable
    //
    //   keyBytes := keyCodec.Append([]byte{}, key)
    //
    keyBytes := keyCodec.Append(nil, key)
    valueBytes, err := EncodeValue(value)
    if err != nil {
        return err
    }
    return db.providerDB.Put(keyBytes, valueBytes)
}

func (db *KeyValueDB) Get(key Key) (*Value, error) {
    keyBytes := keyCodec.Append(nil, key)
    valueBytes, err := db.providerDB.Get(keyBytes)
    if err != nil {
        return nil, err
    }
    return DecodeValue(valueBytes)
}
```

All `Codecs` provided by lexy are safe for concurrent use if their delegate `Codecs` (if any) are.

`Codecs` do not normally encode a data's type, users must know what is being decoded.
This aligns with best practices in Go, types should be known at compile time.
A user-defined `Codec` handling multiple types could be created, but it is not recommended,
and it would still require a concrete wrapper type to conform to the `Codec[T]` interface.

Different `Codecs` will generally not produce encodings with consistent orderings with respect to each other.
For example, the encoding for `int8(1)` will be lexicographically greater than the encoding for `uint8(100)`.

The `Codecs` provided by lexy can encode `nil` to be less than or greater than
the encodings for non-`nil` values, for types that allow `nil` values.

Lexy provides order-preserving `Codecs` for the following types.

* `bool`
* `uint8` (aka `byte`), `uint16`, `uint32`, `uint64`
* `int8`, `int16`, `int32` (aka `rune`), `int64`
* `uint`, `int` (encoded as 64-bit values)
* `float32`, `float64`
* `*math.big.Int`
* `*math.big.Float` (does not encode Accuracy)
* `string`
* `time.Time` (encodes timezone offset, but not its name)
* `time.Duration`
* pointers (also encodes the referent)
* slices
* `[]byte` (optimized for byte slices)

Lexy provides `Codecs` for the following types which either have no natural ordering,
or whose natural ordering cannot be preserved while being encoded at full precision.

* maps
* `complex64`, `complex128`
* `*math.big.Rat`

Lexy provides these additional `Codecs`.

* A `Codec` for types with no value except the zero value, useful for the value types of maps used as sets.
* A `Codec` which reverses the lexicographical ordering of another `Codec`.
* A `Codec` which terminates and escapes the encodings of another `Codec`.

Lexy does not does not provide `Codecs` for the following types, but user-defined `Codecs` are easy to create.
See the Go docs for examples.

* structs  
  The inherent limitations of generic types in Go make it impossible
  to do this in a general way without having a separate parallel set of non-generic codecs.
  This is not a bad thing, resolving types at compile time is one of the reasons Go is so efficient.
  Creating a strongly-typed user-defined `Codec` is a much simpler and safer alternative,
  and also prevents silently changing an encoding when the data type it encodes is changed.
* arrays  
  While it is possible to create a general `Codec` for array types,
  the generics are very messy and it requires using reflection extensively.
  As is the case for structs, creating a strongly-typed user-defined `Codec` is a better option.
* `uintptr`  
  This type has an implementation-specific size,
  and encoding a pointer without encoding what it points to doesn't make much sense.
* functions, interfaces, channels
