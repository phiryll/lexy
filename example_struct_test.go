package lexy_test

import (
	"bytes"
	"fmt"
	"io"
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
	// score sorts high to low
	negScoreCodec = lexy.Negate(lexy.Float32())
	// The type cast is only necessary when using Go versions prior to 1.21.
	tagsCodec       = lexy.TerminateIfNeeded(lexy.Codec[[]string](lexy.SliceOf(lexy.String())))
	SomeStructCodec = someStructCodec{}
)

// Sort order is:
//   - size
//   - score (high to low)
//   - tags
type someStructCodec struct{}

func (c someStructCodec) Read(r io.Reader) (SomeStruct, error) {
	var zero SomeStruct
	size, err := lexy.Int32().Read(r)
	if err != nil {
		return zero, err
	}
	score, err := negScoreCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	tags, err := tagsCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return SomeStruct{size, score, tags}, nil
}

func (c someStructCodec) Write(w io.Writer, value SomeStruct) error {
	if err := lexy.Int32().Write(w, value.size); err != nil {
		return err
	}
	if err := negScoreCodec.Write(w, value.score); err != nil {
		return err
	}
	return tagsCodec.Write(w, value.tags)
}

func (c someStructCodec) RequiresTerminator() bool {
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

var _ sort.Interface = sortableEncodings{}

func (s sortableEncodings) Len() int               { return len(s.b) }
func (s sortableEncodings) Less(i int, j int) bool { return bytes.Compare(s.b[i], s.b[j]) < 0 }
func (s sortableEncodings) Swap(i int, j int)      { s.b[i], s.b[j] = s.b[j], s.b[i] }

// ExampleStruct shows how to define a typical user-defined Codec.
// The rules of thumb are:
//   - The order in which encoded data is written defines the Codec's ordering.
//     Read data in the same order it was written, using the same Codecs.
//     The schema change example has an exception to this.
//   - Use [lexy.ReadPrefix] and [lexy.WritePrefix] if the value can be nil.
//   - Return the type's zero value and [io.EOF] from Read
//     only if no bytes were read and EOF was reached.
//     In the fooCodec example in this comment below,
//     note that the first element read does not use [lexy.UnexpectedIfEOF].
//   - If the value can be nil, lexy.ReadPrefix is the first element read,
//     and the next element read after that should use lexy.UnexpectedIfEOF.
//   - Use [lexy.TerminateIfNeeded] when an element's Codec might require it.
//     See [tagsCodec] in this example for a typical usage.
//   - Return true from [lexy.Codec.RequiresTerminator] when appropriate,
//     whether or not it's relevant at the moment.
//     This allows the Codec to be safely used by others later.
//
// The pattern for a typical struct Codec implementation follows in this comment.
// The same pattern is used for non-struct types,
// see the array example for one such use case.
// The first/second/.../nthCodecs in the example are not meant to imply
// that each field needs its own Codec.
// All the Codecs provided by lexy are safe for concurrent use and reusable,
// assuming their delegate Codecs are safe for concurrent use
// (the argument Codecs used to construct slice and map Codecs, e.g.).
//
//	func (c fooCodec) Read(r io.Reader) (Foo, error) {
//	    var zero Foo
//	    first, err := firstCodec.Read(r)
//	    if err != nil {
//	        return zero, err
//	    }
//	    second, err := secondCodec.Read(r)
//	    if err != nil {
//	        return zero, lexy.UnexpectedIfEOF(err)
//	    }
//	    // ...
//	    nth, err := nthCodec.Read(r)
//	    if err != nil {
//	        return zero, lexy.UnexpectedIfEOF(err)
//	    }
//	    return Foo{first, second, ..., nth}, nil
//	}
//
//	func (c fooCodec) Write(w io.Writer, value Foo) error {
//	    if err := firstCodec.Write(w, value.first); err != nil {
//	        return err
//	    }
//	    if err := secondCodec.Write(w, value.second); err != nil {
//	        return err
//	    }
//	    // ...
//	    return nthCodec.Write(w, value.nth)
//	}
//
//	func (c fooCodec) RequiresTerminator() bool {
//	    return false
//	}
//
// [someStructCodec] in this example code follows the same pattern.
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
	var buf bytes.Buffer

	var encoded [][]byte
	fmt.Println("Round Trip Equals:")
	for _, value := range structs {
		buf.Reset()
		if err := SomeStructCodec.Write(&buf, value); err != nil {
			panic(err)
		}
		encoded = append(encoded, append([]byte{}, buf.Bytes()...))
		decoded, err := SomeStructCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(structsEqual(value, decoded))
	}

	sort.Sort(sortableEncodings{encoded})
	fmt.Println("Sorted:")
	for _, encoded := range encoded {
		decoded, err := SomeStructCodec.Read(bytes.NewReader(encoded))
		if err != nil {
			panic(err)
		}
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
