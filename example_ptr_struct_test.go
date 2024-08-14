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

var (
	PtrToBigStructCodec = ptrToBigStructCodec{}
	ContainerCodec      = containterCodec{}
)

type ptrToBigStructCodec struct{}

func (ptrToBigStructCodec) Read(r io.Reader) (*BigStruct, error) {
	if done, err := lexy.ReadPrefix(r); done {
		return nil, err
	}
	name, err := lexy.TerminatedString().Read(r)
	if err != nil {
		return nil, lexy.UnexpectedIfEOF(err)
	}
	// Read other fields.
	return &BigStruct{name /* , other fields ... */}, nil
}

func (ptrToBigStructCodec) Write(w io.Writer, value *BigStruct) error {
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

func (ptrToBigStructCodec) RequiresTerminator() bool {
	return false
}

type containterCodec struct{}

func (containterCodec) Read(r io.Reader) (Container, error) {
	var zero Container
	big, err := PtrToBigStructCodec.Read(r)
	if err != nil {
		return zero, err
	}
	// Read other fields.
	return Container{big /* , other fields ... */}, nil
}

func (containterCodec) Write(w io.Writer, value Container) error {
	if err := PtrToBigStructCodec.Write(w, value.big); err != nil {
		return err
	}
	// Write other fields.
	return nil
}

func (containterCodec) RequiresTerminator() bool {
	return false
}

// This is only used for printing output in the example.
func containerEquals(a, b Container) bool {
	if a.big == nil && b.big == nil {
		return true
	}
	if a.big == nil || b.big == nil {
		return false
	}
	return *a.big == *b.big
}

// ExamplePointerToStruct shows how to use pointers for efficiency
// in a custom Codec, to avoid unnecessarily copying large data structures.
// Note that types in Go other than structs and arrays do not have this problem.
// Complex numbers, strings, pointers, slices, and maps
// all have a relatively small footprint when passed by value.
// The same is true of [time.Time] and [time.Duration] instances.
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
		fmt.Println(containerEquals(value, decoded))
	}
	// Output:
	// true
	// true
	// true
}
