package lexy_test

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/phiryll/lexy"
)

type schemaVersion1 struct {
	name string
}

type schemaVersion2 struct {
	name     string
	lastName string // added
}

type schemaVersion3 struct {
	name     string
	lastName string
	count    uint16 // added
}

// The current version of the type.
type schemaVersion4 struct {
	firstName  string // renamed from "name"
	middleName string // added
	lastName   string
	// count      uint16 // removed
}

var (
	SchemaVersion1Codec = schemaVersion1Codec{}
	SchemaVersion2Codec = schemaVersion2Codec{}
	SchemaVersion3Codec = schemaVersion3Codec{}
	SchemaVersion4Codec = schemaVersion4Codec{}
	VersionedCodec      = versionedCodec{}
)

type versionedCodec struct{}

func (c versionedCodec) Read(r io.Reader) (schemaVersion4, error) {
	var zero schemaVersion4
	version, err := lexy.Uint[uint32]().Read(r)
	if err != nil {
		return zero, err
	}
	switch version {
	case 1:
		v1, err := lexy.TerminateIfNeeded(SchemaVersion1Codec).Read(r)
		if err != nil {
			return zero, err
		}
		return schemaVersion4{v1.name, "", ""}, nil
	case 2:
		v2, err := lexy.TerminateIfNeeded(SchemaVersion2Codec).Read(r)
		if err != nil {
			return zero, err
		}
		return schemaVersion4{v2.name, "", v2.lastName}, nil
	case 3:
		v3, err := lexy.TerminateIfNeeded(SchemaVersion3Codec).Read(r)
		if err != nil {
			return zero, err
		}
		return schemaVersion4{v3.name, "", v3.lastName}, nil
	case 4:
		return lexy.TerminateIfNeeded(SchemaVersion4Codec).Read(r)
	default:
		panic(fmt.Sprintf("unknown schema version: %d", version))
	}
}

func (c versionedCodec) Write(w io.Writer, value schemaVersion4) error {
	if err := lexy.Uint[uint32]().Write(w, 4); err != nil {
		return err
	}
	return lexy.TerminateIfNeeded(SchemaVersion4Codec).Write(w, value)
}

func (c versionedCodec) RequiresTerminator() bool {
	return false
}

var (
	NameCodec  = lexy.String[string]()
	CountCodec = lexy.Uint[uint16]()
)

// Version 1

type schemaVersion1Codec struct{}

func (p schemaVersion1Codec) Read(r io.Reader) (schemaVersion1, error) {
	var zero schemaVersion1
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	name, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return schemaVersion1{name}, nil
}

func (p schemaVersion1Codec) Write(w io.Writer, value schemaVersion1) error {
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	return terminatedNameCodec.Write(w, value.name)
}

func (p schemaVersion1Codec) RequiresTerminator() bool {
	return false
}

// Version 2

type schemaVersion2Codec struct{}

func (p schemaVersion2Codec) Read(r io.Reader) (schemaVersion2, error) {
	var zero schemaVersion2
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	lastName, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	name, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return schemaVersion2{name, lastName}, nil
}

func (p schemaVersion2Codec) Write(w io.Writer, value schemaVersion2) error {
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	if err := terminatedNameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	return terminatedNameCodec.Write(w, value.name)
}

func (p schemaVersion2Codec) RequiresTerminator() bool {
	return false
}

// Version 3

type schemaVersion3Codec struct{}

func (p schemaVersion3Codec) Read(r io.Reader) (schemaVersion3, error) {
	var zero schemaVersion3
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	count, err := CountCodec.Read(r)
	if err != nil {
		return zero, err
	}
	lastName, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	name, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return schemaVersion3{name, lastName, count}, nil
}

func (p schemaVersion3Codec) Write(w io.Writer, value schemaVersion3) error {
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	if err := CountCodec.Write(w, value.count); err != nil {
		return err
	}
	if err := terminatedNameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	return terminatedNameCodec.Write(w, value.name)
}

func (p schemaVersion3Codec) RequiresTerminator() bool {
	return false
}

// Version 4

type schemaVersion4Codec struct{}

func (s schemaVersion4Codec) Read(r io.Reader) (schemaVersion4, error) {
	var zero schemaVersion4
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	lastName, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	firstName, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	middleName, err := terminatedNameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return schemaVersion4{firstName, middleName, lastName}, nil
}

func (s schemaVersion4Codec) Write(w io.Writer, value schemaVersion4) error {
	terminatedNameCodec := lexy.TerminateIfNeeded(NameCodec)
	if err := terminatedNameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	if err := terminatedNameCodec.Write(w, value.firstName); err != nil {
		return err
	}
	return terminatedNameCodec.Write(w, value.middleName)
}

func (p schemaVersion4Codec) RequiresTerminator() bool {
	return false
}

// A helper function for this test, to write older versions.
func writeWithVersion[T any](w io.Writer, version uint32, codec lexy.Codec[T], value T) error {
	if err := lexy.Uint[uint32]().Write(w, version); err != nil {
		return err
	}
	return lexy.TerminateIfNeeded(codec).Write(w, value)
}

// Example (SchemaVersion) shows how schema versioning could be implemented.
// This can be done in other ways, and more or less leniently.
// This is just an example, and likely a poorly structured one at that.
//
// Note that different encodings of the same type will generally not be ordered
// correctly with respect to each other, regardless of the technique used.
//
// The sort order of encoded data cannot be changed.
// However, there is nothing wrong with creating multiple Codecs
// with different orderings for the same type, nor with storing
// the same data ordered in different ways in the same data store.
func Example_schemaVersion() {
	// Encode data of a bunch of different versions and
	// throw all the encodings into the same slice.
	// Then make sure we can succesfully decode them all.
	var encoded [][]byte
	var buf bytes.Buffer

	// order: name
	for _, v1 := range []schemaVersion1{
		{"Bob"},
		{"Alice"},
		{"Cathy"},
	} {
		buf.Reset()
		if err := writeWithVersion(&buf, 1, SchemaVersion1Codec, v1); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
	}

	// order: lastName, name
	for _, v2 := range []schemaVersion2{
		{"Dave", "Thomas"},
		{"Edgar", "James"},
		{"Fiona", "Smith"},
	} {
		buf.Reset()
		if err := writeWithVersion(&buf, 2, SchemaVersion2Codec, v2); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
	}

	// order: count, lastName, name
	for _, v3 := range []schemaVersion3{
		{"Gloria", "Baker", 6},
		{"Henry", "Washington", 3},
		{"Isabel", "Bardot", 7},
	} {
		buf.Reset()
		if err := writeWithVersion(&buf, 3, SchemaVersion3Codec, v3); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
	}

	// order: lastName, firstName, middleName
	for _, v4 := range []schemaVersion4{
		{"Kevin", "Alex", "Monroe"},
		{"Jennifer", "Anne", "Monroe"},
		{"Lois", "Elizabeth", "Cassidy"},
	} {
		buf.Reset()
		if err := VersionedCodec.Write(&buf, v4); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
	}

	// When the encodings are sorted, they will be in the order:
	// - primary: version
	// - secondary: the encoded order for that version
	slices.SortFunc(encoded, bytes.Compare)

	for _, b := range encoded {
		value, err := VersionedCodec.Read(bytes.NewBuffer(b))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", value)
	}
	// Output:
	// {firstName:Alice middleName: lastName:}
	// {firstName:Bob middleName: lastName:}
	// {firstName:Cathy middleName: lastName:}
	// {firstName:Edgar middleName: lastName:James}
	// {firstName:Fiona middleName: lastName:Smith}
	// {firstName:Dave middleName: lastName:Thomas}
	// {firstName:Henry middleName: lastName:Washington}
	// {firstName:Gloria middleName: lastName:Baker}
	// {firstName:Isabel middleName: lastName:Bardot}
	// {firstName:Lois middleName:Elizabeth lastName:Cassidy}
	// {firstName:Jennifer middleName:Anne lastName:Monroe}
	// {firstName:Kevin middleName:Alex lastName:Monroe}
}
