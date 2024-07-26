package lexy_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/phiryll/lexy"
)

type SimpleStruct struct {
	anInt   int16
	aFloat  float32
	strings []string
}

func (s SimpleStruct) Equals(other SimpleStruct) bool {
	// NaN != NaN, even when they're the exact same bits.
	return s.anInt == other.anInt &&
		math.Float32bits(s.aFloat) == math.Float32bits(other.aFloat) &&
		slices.Equal(s.strings, other.strings)
}

func (s SimpleStruct) String() string {
	return fmt.Sprintf("{%d %.2f %#v}", s.anInt, s.aFloat, s.strings)
}

// Codecs used in these examples.
// All of these are safe for concurrent access.
var (
	anIntCodec      = lexy.Int[int16]()
	aFloatCodec     = lexy.Float32[float32]()
	stringsCodec    = lexy.SliceOf[[]string](lexy.String[string]())
	negStringsCodec = lexy.Negate(stringsCodec)
	ifsCodec        = intFloatStringsCodec{}
	fnsiCodec       = floatNegStringsIntCodec{}
)

// The orderings for these Codecs are in their type names.

type intFloatStringsCodec struct{}

func (c intFloatStringsCodec) Read(r io.Reader) (SimpleStruct, error) {
	var zero SimpleStruct
	anInt, err := anIntCodec.Read(r)
	if err != nil {
		return zero, err
	}
	aFloat, err := aFloatCodec.Read(r)
	if err != nil {
		return zero, err
	}
	strings, err := lexy.TerminateIfNeeded(stringsCodec).Read(r)
	if err != nil {
		return zero, err
	}
	return SimpleStruct{anInt, aFloat, strings}, nil
}

func (c intFloatStringsCodec) Write(w io.Writer, value SimpleStruct) error {
	if err := anIntCodec.Write(w, value.anInt); err != nil {
		return err
	}
	if err := aFloatCodec.Write(w, value.aFloat); err != nil {
		return err
	}
	return lexy.TerminateIfNeeded(stringsCodec).Write(w, value.strings)
}

func (c intFloatStringsCodec) RequiresTerminator() bool {
	return false
}

type floatNegStringsIntCodec struct{}

func (c floatNegStringsIntCodec) Read(r io.Reader) (SimpleStruct, error) {
	var zero SimpleStruct
	aFloat, err := aFloatCodec.Read(r)
	if err != nil {
		return zero, err
	}
	strings, err := lexy.TerminateIfNeeded(negStringsCodec).Read(r)
	if err != nil {
		return zero, err
	}
	anInt, err := anIntCodec.Read(r)
	if err != nil {
		return zero, err
	}
	return SimpleStruct{anInt, aFloat, strings}, nil
}

func (c floatNegStringsIntCodec) Write(w io.Writer, value SimpleStruct) error {
	if err := aFloatCodec.Write(w, value.aFloat); err != nil {
		return err
	}
	if err := lexy.TerminateIfNeeded(negStringsCodec).Write(w, value.strings); err != nil {
		return err
	}
	return anIntCodec.Write(w, value.anInt)
}

func (c floatNegStringsIntCodec) RequiresTerminator() bool {
	return false
}

// Example (SimpleStruct) encodes a struct type using two differently ordered Codecs.
// The pattern will be the same for creating any Codec for a user-defined type.
// Codecs for structs don't usually require enclosing Codecs to use terminators,
// but some do. There are more complex examples in the go docs.
//
// The general rules are:
//   - The order in which encoded data is written defines the Codec's ordering.
//     Read data in the same order it was written, using the same Codecs.
//     An exception to this rule is in the schema change example.
//   - use lexy.Terminate/TerminateIfNeeded for values that do/might
//     require terminating and escaping.
//     It won't be much of a performance hit to use lexy.TerminateIfNeeded,
//     since it returns the argument Codec if it doesn't require termination.
//     These terminating Codecs are not safe for concurrent access,
//     and must be created at the time they are used.
func Example_simpleStruct() {
	structs := []SimpleStruct{
		{1, 5.0, nil},
		{-72, -37.54, []string{"w", "x", "y", "z"}},
		{42, float32(math.Inf(-1)), []string{}},
		{-100, -37.54, []string{"a", "b"}},
		{42, -37.54, []string{"a", "b", "a"}},
		{-100, float32(math.NaN()), []string{"cat"}},
		{42, -37.54, nil},
		{153, -37.54, []string{"d"}},
	}
	var buf bytes.Buffer

	var ifsEncoded [][]byte
	fmt.Println("Int-Float-Strings round trip")
	for _, value := range structs {
		buf.Reset()
		if err := ifsCodec.Write(&buf, value); err != nil {
			panic(err)
		}
		ifsEncoded = append(ifsEncoded, bytes.Clone(buf.Bytes()))
		decoded, err := ifsCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(value.Equals(decoded))
	}

	var fnsiEncoded [][]byte
	fmt.Println("Float-NegStrings-Int round trip")
	for _, value := range structs {
		buf.Reset()
		if err := fnsiCodec.Write(&buf, value); err != nil {
			panic(err)
		}
		fnsiEncoded = append(fnsiEncoded, bytes.Clone(buf.Bytes()))
		decoded, err := fnsiCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(value.Equals(decoded))
	}

	slices.SortFunc(ifsEncoded, bytes.Compare)
	fmt.Println("Int-Float-Strings sorted:")
	for _, encoded := range ifsEncoded {
		decoded, err := ifsCodec.Read(bytes.NewReader(encoded))
		if err != nil {
			panic(err)
		}
		fmt.Println(decoded.String())
	}

	slices.SortFunc(fnsiEncoded, bytes.Compare)
	fmt.Println("Float-NegStrings-Int sorted:")
	for _, encoded := range fnsiEncoded {
		decoded, err := fnsiCodec.Read(bytes.NewReader(encoded))
		if err != nil {
			panic(err)
		}
		fmt.Println(decoded.String())
	}

	// Output:
	// Int-Float-Strings round trip
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// Float-NegStrings-Int round trip
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// Int-Float-Strings sorted:
	// {-100 -37.54 []string{"a", "b"}}
	// {-100 NaN []string{"cat"}}
	// {-72 -37.54 []string{"w", "x", "y", "z"}}
	// {1 5.00 []string(nil)}
	// {42 -Inf []string{}}
	// {42 -37.54 []string(nil)}
	// {42 -37.54 []string{"a", "b", "a"}}
	// {153 -37.54 []string{"d"}}
	// Float-NegStrings-Int sorted:
	// {42 -Inf []string{}}
	// {-72 -37.54 []string{"w", "x", "y", "z"}}
	// {153 -37.54 []string{"d"}}
	// {42 -37.54 []string{"a", "b", "a"}}
	// {-100 -37.54 []string{"a", "b"}}
	// {42 -37.54 []string(nil)}
	// {1 5.00 []string(nil)}
	// {-100 NaN []string{"cat"}}
}
