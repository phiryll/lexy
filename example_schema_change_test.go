package lexy_test

import (
	"fmt"

	"github.com/phiryll/lexy"
)

// The previous version of the type, used here to create already existing data.
// Unless versioning is being used (see the schema version example),
// this would be the same type as schema, just earlier in the code's history.
// So both would not normally exist at the same time.
type schemaPrevious struct {
	name     string
	lastName string
	count    uint16
}

// The current version of the type.
type schema struct {
	firstName  string // renamed from "name"
	middleName string // added
	lastName   string
	// count      uint16 // removed
}

var (
	nameCodec     = lexy.TerminatedString()
	countCodec    = lexy.Uint16()
	PreviousCodec = previousCodec{}
	SchemaCodec   = schemaCodec{}
)

type previousCodec struct{}

func (previousCodec) Append(buf []byte, value schemaPrevious) []byte {
	buf = nameCodec.Append(buf, "count")
	buf = countCodec.Append(buf, value.count)
	buf = nameCodec.Append(buf, "lastName")
	buf = nameCodec.Append(buf, value.lastName)
	buf = nameCodec.Append(buf, "name")
	return nameCodec.Append(buf, value.name)
}

func (previousCodec) Put(_ []byte, _ schemaPrevious) int {
	panic("unused in this example")
}

func (previousCodec) Get(_ []byte) (schemaPrevious, []byte) {
	panic("unused in this example")
}

// Returns true because struct Codecs storing field name/value pairs
// to handle previous versions must be tolerant of missing fields.
// This Codec is essentially a map.
func (previousCodec) RequiresTerminator() bool {
	return true
}

// Other than handling the field changes, this Codec could change the sort order,
// although writing back to the same database index would corrupt the its ordering.
// Because Get reads field names first, it is tolerant of field reorderings.
type schemaCodec struct{}

func (schemaCodec) Append(_ []byte, _ schema) []byte {
	panic("unused in this example")
}

func (schemaCodec) Put(_ []byte, _ schema) int {
	panic("unused in this example")
}

func (schemaCodec) Get(buf []byte) (schema, []byte) {
	var value schema
	for {
		if len(buf) == 0 {
			return value, buf
		}
		var field string
		field, buf = nameCodec.Get(buf)
		switch field {
		case "name", "firstName":
			// Field was renamed.
			firstName, newBuf := nameCodec.Get(buf)
			buf = newBuf
			value.firstName = firstName
		case "middleName":
			// Field was added.
			middleName, newBuf := nameCodec.Get(buf)
			buf = newBuf
			value.middleName = middleName
		case "lastName":
			lastName, newBuf := nameCodec.Get(buf)
			buf = newBuf
			value.lastName = lastName
		case "count":
			// Field was removed, but we still need to read the value.
			_, newBuf := countCodec.Get(buf)
			buf = newBuf
		default:
			// We must stop, we don't know how to proceed.
			panic(fmt.Sprintf("unrecognized field name %q", field))
		}
	}
}

// Returns true because struct Codecs storing field name/value pairs
// to handle previous versions must be tolerant of missing fields.
// This Codec is essentially a map.
func (schemaCodec) RequiresTerminator() bool {
	return true
}

// ExampleSchemaChange shows one way to allow for schema changes.
// The gist of this example is to encode field names as well as field values.
// This can be done in other ways, and more or less leniently.
// This is just an example.
//
// Note that different encodings of the same type will generally not be ordered
// correctly with respect to each other, regardless of the technique used.
//
// Only field values should be encoded if any of the following are true:
//   - the schema is expected to never change, or
//   - the encoded data will be replaced wholesale if the schema changes, or
//   - schema versioning is used (see the schema version example).
//
// The kinds of schema change addressed by this example are:
//   - field added
//   - field removed
//   - field renamed
//
// If a field's type might change, the best option is to use versioning.
// Otherwise, it would be necessary to encode the field's type before its value,
// because there's no way to know how to read the value otherwise,
// and then the type would be the primary sort key for that field.
// Encoding a value's type is strongly discouraged.
//
// The sort order of encoded data cannot be changed.
// However, there is nothing wrong with creating multiple Codecs
// with different orderings for the same type, nor with storing
// the same data ordered in different ways in the same data store.
func Example_schemaChange() {
	for _, previous := range []schemaPrevious{
		{"Alice", "Jones", 35},
		{"", "Washington", 17},
		{"Cathy", "Spencer", 23},
	} {
		buf := PreviousCodec.Append(nil, previous)
		current, _ := SchemaCodec.Get(buf)
		fmt.Println(previous.name == current.firstName &&
			previous.lastName == current.lastName &&
			current.middleName == "")
	}
	// Output:
	// true
	// true
	// true
}
