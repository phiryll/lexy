package lexy_test

import (
	"fmt"
	"sort"

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
	// Which schema this returns will be updated as new versions are added.
	VersionedCodec lexy.Codec[schemaVersion4] = versionedCodec{}

	// The types of the Codecs can be inferred if using Go 1.21 or later.
	SchemaVersion1Codec lexy.Codec[schemaVersion1] = schemaVersion1Codec{}
	SchemaVersion2Codec lexy.Codec[schemaVersion2] = schemaVersion2Codec{}
	SchemaVersion3Codec lexy.Codec[schemaVersion3] = schemaVersion3Codec{}
	SchemaVersion4Codec lexy.Codec[schemaVersion4] = schemaVersion4Codec{}

	NameCodec  = lexy.TerminatedString()
	CountCodec = lexy.Uint16()
)

type versionedCodec struct{}

func (versionedCodec) Append(buf []byte, value schemaVersion4) []byte {
	buf = lexy.Uint32().Append(buf, 4)
	return SchemaVersion4Codec.Append(buf, value)
}

func (versionedCodec) Put(buf []byte, value schemaVersion4) []byte {
	buf = lexy.Uint32().Put(buf, 4)
	return SchemaVersion4Codec.Put(buf, value)
}

func (versionedCodec) Get(buf []byte) (schemaVersion4, []byte) {
	version, buf := lexy.Uint32().Get(buf)
	switch version {
	case 1:
		v1, newBuf := SchemaVersion1Codec.Get(buf)
		return schemaVersion4{v1.name, "", ""}, newBuf
	case 2:
		v2, newBuf := SchemaVersion2Codec.Get(buf)
		return schemaVersion4{v2.name, "", v2.lastName}, newBuf
	case 3:
		v3, newBuf := SchemaVersion3Codec.Get(buf)
		return schemaVersion4{v3.name, "", v3.lastName}, newBuf
	case 4:
		v4, newBuf := SchemaVersion4Codec.Get(buf)
		return v4, newBuf
	default:
		panic(fmt.Sprintf("unknown schema version: %d", version))
	}
}

func (versionedCodec) RequiresTerminator() bool {
	return false
}

// Version 1

type schemaVersion1Codec struct{}

func (schemaVersion1Codec) Append(buf []byte, value schemaVersion1) []byte {
	return NameCodec.Append(buf, value.name)
}

func (schemaVersion1Codec) Put(buf []byte, value schemaVersion1) []byte {
	return NameCodec.Put(buf, value.name)
}

func (schemaVersion1Codec) Get(buf []byte) (schemaVersion1, []byte) {
	name, buf := NameCodec.Get(buf)
	return schemaVersion1{name}, buf
}

func (schemaVersion1Codec) RequiresTerminator() bool {
	return false
}

// Version 2

type schemaVersion2Codec struct{}

func (schemaVersion2Codec) Append(buf []byte, value schemaVersion2) []byte {
	buf = NameCodec.Append(buf, value.lastName)
	return NameCodec.Append(buf, value.name)
}

func (schemaVersion2Codec) Put(buf []byte, value schemaVersion2) []byte {
	buf = NameCodec.Put(buf, value.lastName)
	return NameCodec.Put(buf, value.name)
}

func (schemaVersion2Codec) Get(buf []byte) (schemaVersion2, []byte) {
	lastName, buf := NameCodec.Get(buf)
	name, buf := NameCodec.Get(buf)
	return schemaVersion2{name, lastName}, buf
}

func (schemaVersion2Codec) RequiresTerminator() bool {
	return false
}

// Version 3

type schemaVersion3Codec struct{}

func (schemaVersion3Codec) Append(buf []byte, value schemaVersion3) []byte {
	buf = CountCodec.Append(buf, value.count)
	buf = NameCodec.Append(buf, value.lastName)
	return NameCodec.Append(buf, value.name)
}

func (schemaVersion3Codec) Put(buf []byte, value schemaVersion3) []byte {
	buf = CountCodec.Put(buf, value.count)
	buf = NameCodec.Put(buf, value.lastName)
	return NameCodec.Put(buf, value.name)
}

func (schemaVersion3Codec) Get(buf []byte) (schemaVersion3, []byte) {
	count, buf := CountCodec.Get(buf)
	lastName, buf := NameCodec.Get(buf)
	name, buf := NameCodec.Get(buf)
	return schemaVersion3{name, lastName, count}, buf
}

func (schemaVersion3Codec) RequiresTerminator() bool {
	return false
}

// Version 4

type schemaVersion4Codec struct{}

func (schemaVersion4Codec) Append(buf []byte, value schemaVersion4) []byte {
	buf = NameCodec.Append(buf, value.lastName)
	buf = NameCodec.Append(buf, value.firstName)
	return NameCodec.Append(buf, value.middleName)
}

func (schemaVersion4Codec) Put(buf []byte, value schemaVersion4) []byte {
	buf = NameCodec.Put(buf, value.lastName)
	buf = NameCodec.Put(buf, value.firstName)
	return NameCodec.Put(buf, value.middleName)
}

func (schemaVersion4Codec) Get(buf []byte) (schemaVersion4, []byte) {
	lastName, buf := NameCodec.Get(buf)
	firstName, buf := NameCodec.Get(buf)
	middleName, buf := NameCodec.Get(buf)
	return schemaVersion4{firstName, middleName, lastName}, buf
}

func (schemaVersion4Codec) RequiresTerminator() bool {
	return false
}

// A helper function for this test, to write older versions.
func writeWithVersion[T any](version uint32, codec lexy.Codec[T], value T) []byte {
	buf := lexy.Uint32().Append(nil, version)
	return codec.Append(buf, value)
}

// ExampleSchemaVersion shows how schema versioning could be implemented.
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
	// Then make sure we can successfully decode them all.
	var encoded [][]byte

	// order: name
	for _, v1 := range []schemaVersion1{
		{"Bob"},
		{"Alice"},
		{"Cathy"},
	} {
		encoded = append(encoded, writeWithVersion(1, SchemaVersion1Codec, v1))
	}

	// order: lastName, name
	for _, v2 := range []schemaVersion2{
		{"Dave", "Thomas"},
		{"Edgar", "James"},
		{"Fiona", "Smith"},
	} {
		encoded = append(encoded, writeWithVersion(2, SchemaVersion2Codec, v2))
	}

	// order: count, lastName, name
	for _, v3 := range []schemaVersion3{
		{"Gloria", "Baker", 6},
		{"Henry", "Washington", 3},
		{"Isabel", "Bardot", 7},
	} {
		encoded = append(encoded, writeWithVersion(3, SchemaVersion3Codec, v3))
	}

	// order: lastName, firstName, middleName
	for _, v4 := range []schemaVersion4{
		{"Kevin", "Alex", "Monroe"},
		{"Jennifer", "Anne", "Monroe"},
		{"Lois", "Elizabeth", "Cassidy"},
	} {
		encoded = append(encoded, VersionedCodec.Append(nil, v4))
	}

	// When the encodings are sorted, they will be in the order:
	// - primary: version
	// - secondary: the encoded order for that version
	// sortableEncodings is defined in the Struct example.
	sort.Sort(sortableEncodings{encoded})

	for _, b := range encoded {
		value, _ := VersionedCodec.Get(b)
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
