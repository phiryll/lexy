package lexy_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/phiryll/lexy"
)

type BigStruct struct {
	name string
	// ... big fields, inefficient to copy
}

type Container struct {
	big *BigStruct
	// ... other fields, but not large ones
	// If Container were also large, use the same technique
	// to create a Codec[*Container] instead.
}

func (c Container) Equals(other Container) bool {
	if c.big == nil && other.big == nil {
		return true
	}
	if c.big == nil || other.big == nil {
		return false
	}
	return *c.big == *other.big
}

var (
	PtrToBigStructCodec = ptrToBigStructCodec{}
	ContainerCodec      = containterCodec{}
)

// A Codec[*BigStruct]
type ptrToBigStructCodec struct{}

func (c ptrToBigStructCodec) Read(r io.Reader) (*BigStruct, error) {
	if done, err := lexy.ReadPrefix(r); done {
		return nil, err
	}
	name, err := lexy.TerminatedString().Read(r)
	if err != nil {
		return nil, lexy.UnexpectedIfEOF(err)
	}
	// Read other fields.
	return &BigStruct{name /*, other fields ...*/}, nil
}

func (c ptrToBigStructCodec) Write(w io.Writer, value *BigStruct) error {
	// done is true if there was an error, or if value is nil,
	// in which case a prefix denoting "nil" has already been written.
	if done, err := lexy.WritePrefix(w, value == nil, true); done {
		return err
	}
	if err := lexy.TerminatedString().Write(w, value.name); err != nil {
		return err
	}
	// Write other fields.
	return nil
}

// A Codec[Container]
func (c ptrToBigStructCodec) RequiresTerminator() bool {
	return false
}

type containterCodec struct{}

func (c containterCodec) Read(r io.Reader) (Container, error) {
	big, err := PtrToBigStructCodec.Read(r)
	if err != nil {
		var zero Container
		return zero, err
	}
	// Read other fields.
	return Container{big /*, other fields ...*/}, nil
}

func (c containterCodec) Write(w io.Writer, value Container) error {
	if err := PtrToBigStructCodec.Write(w, value.big); err != nil {
		return err
	}
	// Write other fields.
	return nil
}

func (c containterCodec) RequiresTerminator() bool {
	return false
}

// Example (PointerToStruct) shows how to use pointers for efficiency
// in a custom Codec, to avoid unnecessarily copying large data structures.
// Note that types in go other than structs and arrays do not have this problem.
// Complex numbers, strings, pointers, slices, and maps
// all have a relatively small footprint when passed by value.
// The same is true of time.Time and time.Duration instances.
//
// Normally, a Codec[BigStruct] would be defined and Container's Codec
// would use it as lexy.PointerTo(bigStructCodec).
// However, calls to a Codec[BigStruct] will pass BigStruct instances by value,
// even though the wrapping pointer Codec is only copying pointers.
//
// The order isn't relevant for this example, so other fields are not shown.
func Example_pointerToStruct() {
	var buf bytes.Buffer
	for _, value := range []Container{
		{nil},
		{&BigStruct{""}},
		{&BigStruct{"abc"}},
	} {
		buf.Reset()
		if err := ContainerCodec.Write(&buf, value); err != nil {
			panic(err)
		}
		decoded, err := ContainerCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(value.Equals(decoded))
	}
	// Output:
	// true
	// true
	// true
}
