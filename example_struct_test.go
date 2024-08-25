package lexy_test

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/phiryll/lexy"
)

type SomeStruct struct {
	size  int32
	score float32
	tags  []string
}

func (s SomeStruct) String() string {
	return fmt.Sprintf("{%d %.2f %#v}", s.size, s.score, s.tags)
}

// All of these are safe for concurrent access.
var (
	// Score sorts high to low.
	negScoreCodec   = lexy.Negate(lexy.Float32())
	tagsCodec       = lexy.TerminateIfNeeded(lexy.SliceOf(lexy.String()))
	SomeStructCodec = someStructCodec{}
)

// Sort order is:
//   - size
//   - score (high to low)
//   - tags
type someStructCodec struct{}

func (someStructCodec) Append(buf []byte, value SomeStruct) []byte {
	buf = lexy.Int32().Append(buf, value.size)
	buf = negScoreCodec.Append(buf, value.score)
	return tagsCodec.Append(buf, value.tags)
}

func (someStructCodec) Put(buf []byte, value SomeStruct) []byte {
	buf = lexy.Int32().Put(buf, value.size)
	buf = negScoreCodec.Put(buf, value.score)
	return tagsCodec.Put(buf, value.tags)
}

func (someStructCodec) Get(buf []byte) (SomeStruct, []byte) {
	size, buf := lexy.Int32().Get(buf)
	score, buf := negScoreCodec.Get(buf)
	tags, buf := tagsCodec.Get(buf)
	return SomeStruct{size, score, tags}, buf
}

func (someStructCodec) RequiresTerminator() bool {
	return false
}

// Only defined to test whether two SomeStructs are equal.
func structsEqual(a, b SomeStruct) bool {
	if a.size != b.size {
		return false
	}
	// NaN != NaN, even when they're the exact same bits.
	if math.Float32bits(a.score) != math.Float32bits(b.score) {
		return false
	}
	if len(a.tags) != len(b.tags) {
		return false
	}
	for i := range a.tags {
		if a.tags[i] != b.tags[i] {
			return false
		}
	}
	return true
}

type sortableEncodings struct {
	b [][]byte
}

var _ sort.Interface = sortableEncodings{nil}

func (s sortableEncodings) Len() int           { return len(s.b) }
func (s sortableEncodings) Less(i, j int) bool { return bytes.Compare(s.b[i], s.b[j]) < 0 }
func (s sortableEncodings) Swap(i, j int)      { s.b[i], s.b[j] = s.b[j], s.b[i] }

// ExampleStruct shows how to define a typical user-defined Codec.
// someStructCodec in this example demonstrates an idiomatic Codec definition.
// The same pattern is used for non-struct types,
// see the array example for one such use case.
//
// The rules of thumb are:
//   - The order in which encoded data is written defines the Codec's ordering.
//     Get should read data in the same order it was written, using the same Codecs.
//     The schema change example has an exception to this.
//   - Use [lexy.PrefixNilsFirst] or [lexy.PrefixNilsLast] if the value can be nil.
//   - Get must panic if it cannot decode a value.
//   - Use [lexy.TerminateIfNeeded] when an element's Codec might require it.
//     See [tagsCodec] in this example for a typical usage.
//   - Return true from [lexy.Codec.RequiresTerminator] when appropriate,
//     whether or not it's relevant at the moment.
//     This allows the Codec to be safely used by others later.
func Example_struct() {
	structs := []SomeStruct{
		{1, 5.0, nil},
		{-72, 37.54, []string{"w", "x", "y", "z"}},
		{42, 37.6, []string{"p", "q", "r"}},
		{42, float32(math.Inf(1)), []string{}},
		{-100, 37.54, []string{"a", "b"}},
		{42, 37.54, []string{"a", "b", "a"}},
		{-100, float32(math.NaN()), []string{"cat"}},
		{42, 37.54, nil},
		{153, 37.54, []string{"d"}},
	}

	var encoded [][]byte
	fmt.Println("Round Trip Equals:")
	for _, value := range structs {
		buf := SomeStructCodec.Append(nil, value)
		decoded, _ := SomeStructCodec.Get(buf)
		fmt.Println(structsEqual(value, decoded))
		encoded = append(encoded, buf)
	}

	sort.Sort(sortableEncodings{encoded})
	fmt.Println("Sorted:")
	for _, enc := range encoded {
		decoded, _ := SomeStructCodec.Get(enc)
		fmt.Println(decoded.String())
	}

	// Output:
	// Round Trip Equals:
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// Sorted:
	// {-100 NaN []string{"cat"}}
	// {-100 37.54 []string{"a", "b"}}
	// {-72 37.54 []string{"w", "x", "y", "z"}}
	// {1 5.00 []string(nil)}
	// {42 +Inf []string{}}
	// {42 37.60 []string{"p", "q", "r"}}
	// {42 37.54 []string(nil)}
	// {42 37.54 []string{"a", "b", "a"}}
	// {153 37.54 []string{"d"}}
}
