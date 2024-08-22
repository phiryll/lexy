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
	done, newBuf := lexy.PrefixNilsFirst.Append(buf, value == nil)
	if done {
		return newBuf
	}
	// Append other fields.
	return lexy.TerminatedString().Append(newBuf, value.name)
}

func (ptrToBigStructCodec) Put(buf []byte, value *BigStruct) int {
	if lexy.PrefixNilsFirst.Put(buf, value == nil) {
		return 1
	}
	n := 1
	// Put other fields.
	n += lexy.TerminatedString().Put(buf[n:], value.name)
	return n
}

func (ptrToBigStructCodec) Get(buf []byte) (*BigStruct, int) {
	if len(buf) == 0 {
		return nil, -1
	}
	if lexy.PrefixNilsFirst.Get(buf) {
		return nil, 1
	}
	n := 1
	name, count := lexy.TerminatedString().Get(buf[n:])
	n += count
	// Get other fields.
	return &BigStruct{name /* , other fields ... */}, n
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

func (containterCodec) Put(buf []byte, value Container) int {
	n := PtrToBigStructCodec.Put(buf, value.big)
	// Put other fields.
	// n += someCodec.Put(buf[n:], someValue)
	return n
}

func (containterCodec) Get(buf []byte) (Container, int) {
	big, n := PtrToBigStructCodec.Get(buf)
	// Get other fields.
	// someValue, count := someCodec.Get(buf[n:])
	// n += count
	return Container{big /* , other fields ... */}, n
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
