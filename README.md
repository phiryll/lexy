# lexy

[![Build Status](https://github.com/phiryll/lexy/actions/workflows/tests.yaml/badge.svg?branch=main)](https://github.com/phiryll/lexy/actions/workflows/tests.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/phiryll/lexy)](https://goreportcard.com/report/github.com/phiryll/lexy)
[![Go Reference](https://pkg.go.dev/badge/github.com/phiryll/lexy)](https://pkg.go.dev/github.com/phiryll/lexy)

Lexicographical Byte Order Encodings

Lexy is a library for order-preserving lexicographical binary encodings.
Most common Go types and user-defined types are supported,
and it allows for encodings ordered differently than a type's natural ordering.
Lexy uses generics and requires Go 1.18 to use. It has been tested through Go 1.22.
Lexy has no non-test dependencies.

It may be more efficient to use another encoding if lexicographical unsigned byte ordering is not needed.
Lexy's primary purpose is to make it easier to use an
[ordered key-value store](https://en.wikipedia.org/wiki/Ordered_Key-Value_Store),
an ordered binary trie, or similar.

The primary interface in lexy is `Codec`, with this definition (details in Go docs):

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
var keyCodec = lexy.MakeSliceOf[Key](lexy.MakeString[Word]())

// The terser functions lexy.SliceOf and lexy.String can be used
// if the types involved are the same as their underlying types,
// string and []string in this case. That would look like this:
//
// var keyCodec = lexy.SliceOf(lexy.String())

// lexy could be used here, but it's overkill if ordered Values aren't needed.
func EncodeValue(v *Value) ([]byte, error) { /* ... */ }
func DecodeValue(b []byte) (*Value, error) { /* ... */ }

type KeyValueDB struct {
    providerDB *provider.DB
    // ...
}

func (db *KeyValueDB) Put(key Key, value *Value) error {
    keyBytes, err := lexy.Encode(keyCodec, key)
    if err != nil {
        return err
    }
    valueBytes, err := EncodeValue(value)
    if err != nil {
        return err
    }
    return db.providerDB.Put(keyBytes, valueBytes)
}

func (db *KeyValueDB) Get(key Key) (*Value, error) {
    keyBytes, err := lexy.Encode(keyCodec, key)
    if err != nil {
        return nil, err
    }
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
A custom `Codec` handling multiple types could be created, but it is not recommended,
and it would still require a concrete wrapper type to conform to the `Codec[T]` interface.

Different `Codecs` will generally not produce encodings with consistent orderings with respect to each other.
For example, `Encode(Int8(), 1)` will produce an encoding
that is lexicographically greater than the encoding produced by `Encode(Uint8(), 100)`.

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

Lexy provides a `Codec` for types with no value except the zero value,
useful for the value types of maps used as sets.

Lexy does not does not provide `Codecs` for the following types, but custom `Codecs` are easy to create.
See the Go docs for examples.

* structs  
  The inherent limitations of generic types in Go make it impossible
  to do this in a general way without having a separate parallel set of non-generic codecs.
  This is not a bad thing, resolving types at compile time is one of the reasons Go is so efficient.
  Creating a strongly-typed custom `Codec` is a much simpler and safer alternative,
  and also prevents silently changing an encoding when the data type it encodes is changed.
* arrays  
  While it is possible to create a general `Codec` for array types,
  the generics are very messy and it requires using reflection extensively.
  As is the case for structs, creating a strongly-typed custom `Codec` is a better option.
* `uintptr`  
  This type has an implementation-specific size,
  and encoding a pointer without encoding what it points to doesn't make much sense.
* functions, interfaces, channels
