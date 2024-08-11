package lexy_test

import (
	"bytes"
	"fmt"
	"io"

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

func (p previousCodec) Read(_ io.Reader) (schemaPrevious, error) {
	panic("unused in this example")
}

func (p previousCodec) Write(w io.Writer, value schemaPrevious) error {
	if err := nameCodec.Write(w, "count"); err != nil {
		return err
	}
	if err := countCodec.Write(w, value.count); err != nil {
		return err
	}
	if err := nameCodec.Write(w, "lastName"); err != nil {
		return err
	}
	if err := nameCodec.Write(w, value.lastName); err != nil {
		return err
	}
	if err := nameCodec.Write(w, "name"); err != nil {
		return err
	}
	return nameCodec.Write(w, value.name)
}

// Returns true because struct Codecs storing field name/value pairs
// to handle previous versions must be tolerant of missing fields.
// This Codec is essentially a map.
func (p previousCodec) RequiresTerminator() bool {
	return true
}

// Other than handling the field changes, this Codec could change the sort order.
// Because Read reads field names first, it is tolerant of field reorderings.
type schemaCodec struct{}

func (s schemaCodec) Read(r io.Reader) (schema, error) {
	var zero, value schema
	for {
		field, err := nameCodec.Read(r)
		if err == io.EOF {
			// EOF at this point means we're done.
			return value, nil
		}
		if err != nil {
			return zero, err
		}
		switch field {
		case "name", "firstName":
			// Field was renamed.
			firstName, err := nameCodec.Read(r)
			if err != nil {
				return zero, lexy.UnexpectedIfEOF(err)
			}
			value.firstName = firstName
		case "middleName":
			// Field was added.
			middleName, err := nameCodec.Read(r)
			if err != nil {
				return zero, lexy.UnexpectedIfEOF(err)
			}
			value.middleName = middleName
		case "lastName":
			lastName, err := nameCodec.Read(r)
			if err != nil {
				return zero, lexy.UnexpectedIfEOF(err)
			}
			value.lastName = lastName
		case "count":
			// Field was removed, but we still need to read the value.
			_, err := countCodec.Read(r)
			if err != nil {
				return zero, lexy.UnexpectedIfEOF(err)
			}
		default:
			// We must stop, we don't know how to proceed.
			panic(fmt.Sprintf("unrecognized field name %q", field))
		}
	}
}

func (s schemaCodec) Write(_ io.Writer, _ schema) error {
	panic("unused in this example")
}

// Returns true because struct Codecs storing field name/value pairs
// to handle previous versions must be tolerant of missing fields.
// This Codec is essentially a map.
func (s schemaCodec) RequiresTerminator() bool {
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
// Encoding a value's type is discouraged.
//
// The sort order of encoded data cannot be changed.
// However, there is nothing wrong with creating multiple Codecs
// with different orderings for the same type, nor with storing
// the same data ordered in different ways in the same data store.
func Example_schemaChange() {
	var buf bytes.Buffer
	for _, previous := range []schemaPrevious{
		{"Alice", "Jones", 35},
		{"", "Washington", 17},
		{"Cathy", "Spencer", 23},
	} {
		buf.Reset()
		if err := PreviousCodec.Write(&buf, previous); err != nil {
			panic(err)
		}
		current, err := SchemaCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(previous.name == current.firstName &&
			previous.lastName == current.lastName &&
			current.middleName == "")
	}
	// Output:
	// true
	// true
	// true
}
