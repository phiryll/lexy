package lexy_test

import (
	"fmt"

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
	PtrToBigStructCodec lexy.Codec[*BigStruct] = ptrToBigStructCodec{}
	ContainerCodec      lexy.Codec[Container]  = containterCodec{}
)

type ptrToBigStructCodec struct{}

func (ptrToBigStructCodec) Append(buf []byte, value *BigStruct) []byte {
	done, buf := lexy.PrefixNilsFirst.Append(buf, value == nil)
	if done {
		return buf
	}
	buf = lexy.TerminatedString().Append(buf, value.name)
	// Append other fields.
	return buf
}

func (ptrToBigStructCodec) Put(buf []byte, value *BigStruct) []byte {
	done, buf := lexy.PrefixNilsFirst.Put(buf, value == nil)
	if done {
		return buf
	}
	buf = lexy.TerminatedString().Put(buf, value.name)
	// Put other fields.
	return buf
}

func (ptrToBigStructCodec) Get(buf []byte) (*BigStruct, []byte) {
	done, buf := lexy.PrefixNilsFirst.Get(buf)
	if done {
		return nil, buf
	}
	name, buf := lexy.TerminatedString().Get(buf)
	// Get other fields.
	return &BigStruct{name /* , other fields ... */}, buf
}

func (ptrToBigStructCodec) RequiresTerminator() bool {
	return false
}

type containterCodec struct{}

func (containterCodec) Append(buf []byte, value Container) []byte {
	buf = PtrToBigStructCodec.Append(buf, value.big)
	// Append other fields.
	return buf
}

func (containterCodec) Put(buf []byte, value Container) []byte {
	buf = PtrToBigStructCodec.Put(buf, value.big)
	// Put other fields.
	// buf = someCodec.Put(buf, someValue)
	return buf
}

func (containterCodec) Get(buf []byte) (Container, []byte) {
	big, buf := PtrToBigStructCodec.Get(buf)
	// Get other fields.
	// someValue, buf := someCodec.Get(buf)
	// ...
	return Container{big /* , other fields ... */}, buf
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
// in a user-defined Codec, to avoid unnecessarily copying large data structures.
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
	for _, value := range []Container{
		{nil},
		{&BigStruct{""}},
		{&BigStruct{"abc"}},
	} {
		buf := ContainerCodec.Append(nil, value)
		decoded, _ := ContainerCodec.Get(buf)
		fmt.Println(containerEquals(value, decoded))
	}
	// Output:
	// true
	// true
	// true
}
