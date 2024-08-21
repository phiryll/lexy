package lexy_test

import (
	"bytes"
	"fmt"
	"io"
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
	// The types of the Codecs can be inferred if using Go 1.21 or later.
	SchemaVersion1Codec lexy.Codec[schemaVersion1] = schemaVersion1Codec{}
	SchemaVersion2Codec lexy.Codec[schemaVersion2] = schemaVersion2Codec{}
	SchemaVersion3Codec lexy.Codec[schemaVersion3] = schemaVersion3Codec{}
	SchemaVersion4Codec lexy.Codec[schemaVersion4] = schemaVersion4Codec{}

	// Which schema this returns will be updated as new versions are added.
	VersionedCodec lexy.Codec[schemaVersion4] = versionedCodec{}

	NameCodec  = lexy.TerminatedString()
	CountCodec = lexy.Uint16()
)

type versionedCodec struct{}

func (c versionedCodec) Append(buf []byte, value schemaVersion4) []byte {
	return lexy.AppendUsingWrite[schemaVersion4](c, buf, value)
}

func (c versionedCodec) Put(buf []byte, value schemaVersion4) int {
	return lexy.PutUsingAppend[schemaVersion4](c, buf, value)
}

func (c versionedCodec) Get(buf []byte) (schemaVersion4, int) {
	return lexy.GetUsingRead[schemaVersion4](c, buf)
}

func (versionedCodec) Write(w io.Writer, value schemaVersion4) error {
	if err := lexy.Uint32().Write(w, 4); err != nil {
		return err
	}
	return SchemaVersion4Codec.Write(w, value)
}

func (versionedCodec) Read(r io.Reader) (schemaVersion4, error) {
	var zero schemaVersion4
	version, err := lexy.Uint32().Read(r)
	if err != nil {
		return zero, err
	}
	switch version {
	case 1:
		v1, err := SchemaVersion1Codec.Read(r)
		if err != nil {
			return zero, lexy.UnexpectedIfEOF(err)
		}
		return schemaVersion4{v1.name, "", ""}, nil
	case 2:
		v2, err := SchemaVersion2Codec.Read(r)
		if err != nil {
			return zero, lexy.UnexpectedIfEOF(err)
		}
		return schemaVersion4{v2.name, "", v2.lastName}, nil
	case 3:
		v3, err := SchemaVersion3Codec.Read(r)
		if err != nil {
			return zero, lexy.UnexpectedIfEOF(err)
		}
		return schemaVersion4{v3.name, "", v3.lastName}, nil
	case 4:
		v4, err := SchemaVersion4Codec.Read(r)
		if err != nil {
			return zero, lexy.UnexpectedIfEOF(err)
		}
		return v4, nil
	default:
		panic(fmt.Sprintf("unknown schema version: %d", version))
	}
}

func (versionedCodec) RequiresTerminator() bool {
	return false
}

// Version 1

type schemaVersion1Codec struct{}

func (c schemaVersion1Codec) Append(buf []byte, value schemaVersion1) []byte {
	return lexy.AppendUsingWrite[schemaVersion1](c, buf, value)
}

func (c schemaVersion1Codec) Put(buf []byte, value schemaVersion1) int {
	return lexy.PutUsingAppend[schemaVersion1](c, buf, value)
}

func (c schemaVersion1Codec) Get(buf []byte) (schemaVersion1, int) {
	return lexy.GetUsingRead[schemaVersion1](c, buf)
}

func (schemaVersion1Codec) Write(w io.Writer, value schemaVersion1) error {
	return NameCodec.Write(w, value.name)
}

func (schemaVersion1Codec) Read(r io.Reader) (schemaVersion1, error) {
	var zero schemaVersion1
	name, err := NameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return schemaVersion1{name}, nil
}

func (schemaVersion1Codec) RequiresTerminator() bool {
	return false
}

// Version 2

type schemaVersion2Codec struct{}

func (c schemaVersion2Codec) Append(buf []byte, value schemaVersion2) []byte {
	return lexy.AppendUsingWrite[schemaVersion2](c, buf, value)
}

func (c schemaVersion2Codec) Put(buf []byte, value schemaVersion2) int {
	return lexy.PutUsingAppend[schemaVersion2](c, buf, value)
}

func (c schemaVersion2Codec) Get(buf []byte) (schemaVersion2, int) {
	return lexy.GetUsingRead[schemaVersion2](c, buf)
}

func (schemaVersion2Codec) Write(w io.Writer, value schemaVersion2) error {
	if err := NameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	return NameCodec.Write(w, value.name)
}

func (schemaVersion2Codec) Read(r io.Reader) (schemaVersion2, error) {
	var zero schemaVersion2
	lastName, err := NameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	name, err := NameCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return schemaVersion2{name, lastName}, nil
}

func (schemaVersion2Codec) RequiresTerminator() bool {
	return false
}

// Version 3

type schemaVersion3Codec struct{}

func (c schemaVersion3Codec) Append(buf []byte, value schemaVersion3) []byte {
	return lexy.AppendUsingWrite[schemaVersion3](c, buf, value)
}

func (c schemaVersion3Codec) Put(buf []byte, value schemaVersion3) int {
	return lexy.PutUsingAppend[schemaVersion3](c, buf, value)
}

func (c schemaVersion3Codec) Get(buf []byte) (schemaVersion3, int) {
	return lexy.GetUsingRead[schemaVersion3](c, buf)
}

func (schemaVersion3Codec) Write(w io.Writer, value schemaVersion3) error {
	if err := CountCodec.Write(w, value.count); err != nil {
		return err
	}
	if err := NameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	return NameCodec.Write(w, value.name)
}

func (schemaVersion3Codec) Read(r io.Reader) (schemaVersion3, error) {
	var zero schemaVersion3
	count, err := CountCodec.Read(r)
	if err != nil {
		return zero, err
	}
	lastName, err := NameCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	name, err := NameCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return schemaVersion3{name, lastName, count}, nil
}

func (schemaVersion3Codec) RequiresTerminator() bool {
	return false
}

// Version 4

type schemaVersion4Codec struct{}

func (c schemaVersion4Codec) Append(buf []byte, value schemaVersion4) []byte {
	return lexy.AppendUsingWrite[schemaVersion4](c, buf, value)
}

func (c schemaVersion4Codec) Put(buf []byte, value schemaVersion4) int {
	return lexy.PutUsingAppend[schemaVersion4](c, buf, value)
}

func (c schemaVersion4Codec) Get(buf []byte) (schemaVersion4, int) {
	return lexy.GetUsingRead[schemaVersion4](c, buf)
}

func (schemaVersion4Codec) Write(w io.Writer, value schemaVersion4) error {
	if err := NameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	if err := NameCodec.Write(w, value.firstName); err != nil {
		return err
	}
	return NameCodec.Write(w, value.middleName)
}

func (schemaVersion4Codec) Read(r io.Reader) (schemaVersion4, error) {
	var zero schemaVersion4
	lastName, err := NameCodec.Read(r)
	if err != nil {
		return zero, err
	}
	firstName, err := NameCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	middleName, err := NameCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return schemaVersion4{firstName, middleName, lastName}, nil
}

func (schemaVersion4Codec) RequiresTerminator() bool {
	return false
}

// A helper function for this test, to write older versions.
func writeWithVersion[T any](w io.Writer, version uint32, codec lexy.Codec[T], value T) error {
	if err := lexy.Uint32().Write(w, version); err != nil {
		return err
	}
	return codec.Write(w, value)
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
		encoded = append(encoded, append([]byte{}, buf.Bytes()...))
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
		encoded = append(encoded, append([]byte{}, buf.Bytes()...))
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
		encoded = append(encoded, append([]byte{}, buf.Bytes()...))
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
		encoded = append(encoded, append([]byte{}, buf.Bytes()...))
	}

	// When the encodings are sorted, they will be in the order:
	// - primary: version
	// - secondary: the encoded order for that version
	// sortableEncodings is defined in the Struct example.
	sort.Sort(sortableEncodings{encoded})

	for _, b := range encoded {
		value, err := VersionedCodec.Read(bytes.NewReader(b))
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
