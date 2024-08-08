package lexy_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/phiryll/lexy"
)

type SimpleStruct struct {
	anInt   int16
	aFloat  float32
	strings []string
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s SimpleStruct) Equals(other SimpleStruct) bool {
	// NaN != NaN, even when they're the exact same bits.
	return s.anInt == other.anInt &&
		math.Float32bits(s.aFloat) == math.Float32bits(other.aFloat) &&
		stringsEqual(s.strings, other.strings)
}

func (s SimpleStruct) String() string {
	return fmt.Sprintf("{%d %.2f %#v}", s.anInt, s.aFloat, s.strings)
}

// Codecs used in these examples.
// All of these are safe for concurrent access.
var (
	anIntCodec  = lexy.Int16()
	aFloatCodec = lexy.Float32()
	// The cast is only necessary when using go versions prior to 1.21.
	stringsCodec    = lexy.Terminate(lexy.Codec[[]string](lexy.SliceOf(lexy.String())))
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
		return zero, lexy.UnexpectedIfEOF(err)
	}
	strings, err := stringsCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
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
	return stringsCodec.Write(w, value.strings)
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
	strings, err := negStringsCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	anInt, err := anIntCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return SimpleStruct{anInt, aFloat, strings}, nil
}

func (c floatNegStringsIntCodec) Write(w io.Writer, value SimpleStruct) error {
	if err := aFloatCodec.Write(w, value.aFloat); err != nil {
		return err
	}
	if err := negStringsCodec.Write(w, value.strings); err != nil {
		return err
	}
	return anIntCodec.Write(w, value.anInt)
}

func (c floatNegStringsIntCodec) RequiresTerminator() bool {
	return false
}

type sortableWrapper struct {
	b [][]byte
}

var _ sort.Interface = sortableWrapper{}

func (s sortableWrapper) Len() int               { return len(s.b) }
func (s sortableWrapper) Less(i int, j int) bool { return bytes.Compare(s.b[i], s.b[j]) < 0 }
func (s sortableWrapper) Swap(i int, j int)      { s.b[i], s.b[j] = s.b[j], s.b[i] }

// Example (SimpleStruct) encodes a struct type using two differently ordered Codecs.
// The pattern will be the same for creating any Codec for a user-defined type.
// Codecs for structs don't usually require enclosing Codecs to use terminators,
// but some do. There are more complex examples in the go docs.
//
// The general rules are:
//   - The order in which encoded data is written defines the Codec's ordering.
//     Read data in the same order it was written, using the same Codecs.
//     An exception to this rule is in the schema change example.
//   - Use lexy.Terminate/TerminateIfNeeded for values that do/might
//     require terminating and escaping.
//     It won't be much of a performance hit to use lexy.TerminateIfNeeded,
//     since it returns the argument Codec if it doesn't require termination.
//   - If no bytes are read and EOF is reached when reading,
//     return the zero value and io.EOF from Codec.Read.
//     If EOF is clearly reached before a complete value has been read,
//     return io.ErrUnexpectedEOF.
//     If EOF is reached and what has been read could be a complete value,
//     return the value and no error.
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
		ifsEncoded = append(ifsEncoded, append([]byte{}, buf.Bytes()...))
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
		fnsiEncoded = append(fnsiEncoded, append([]byte{}, buf.Bytes()...))
		decoded, err := fnsiCodec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(value.Equals(decoded))
	}

	sort.Sort(sortableWrapper{ifsEncoded})
	fmt.Println("Int-Float-Strings sorted:")
	for _, encoded := range ifsEncoded {
		decoded, err := ifsCodec.Read(bytes.NewReader(encoded))
		if err != nil {
			panic(err)
		}
		fmt.Println(decoded.String())
	}

	sort.Sort(sortableWrapper{fnsiEncoded})
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
