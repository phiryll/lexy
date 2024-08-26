package lexy_test

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/phiryll/lexy"
)

func ExampleEmpty() {
	type present struct{}
	type set map[uint8]present
	codec := lexy.CastMapOf[set](lexy.Uint8(), lexy.Empty[present]())
	value := set{
		23: present{},
		42: present{},
		59: present{},
		12: present{},
	}
	buf := codec.Append(nil, value)
	decoded, _ := codec.Get(buf)
	fmt.Printf("%T\n", decoded)
	fmt.Printf("%T\n", decoded[0])
	fmt.Println(reflect.DeepEqual(value, decoded))
	// Output:
	// lexy_test.set
	// lexy_test.present
	// true
}

func ExampleBool() {
	codec := lexy.Bool()
	buf := codec.Append(nil, true)
	first, _ := codec.Get(buf)
	_ = codec.Put(buf, false)
	second, _ := codec.Get(buf)
	fmt.Printf("%t, %t", first, second)
	// Output:
	// true, false
}

func ExampleUint64() {
	codec := lexy.Uint64()
	buf := codec.Append(nil, 123)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// 123
}

func ExampleUint() {
	codec := lexy.Uint()
	buf := codec.Append(nil, 4567890)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// 4567890
}

func ExampleUint32_underlyingType() {
	type size uint32
	codec := lexy.CastUint32[size]()
	buf := codec.Append(nil, 123)
	decoded, _ := codec.Get(buf)
	fmt.Printf("Value %d of type %T", decoded, decoded)
	// Output:
	// Value 123 of type lexy_test.size
}

func ExampleInt32() {
	codec := lexy.Int32()
	var encoded [][]byte
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		buf := codec.Append(nil, value)
		encoded = append(encoded, buf)
	}
	// Verify the encodings are increasing.
	for i, b := range encoded[1:] {
		fmt.Println(bytes.Compare(encoded[i], b))
	}
	// Output:
	// -1
	// -1
	// -1
	// -1
}

func ExampleInt() {
	codec := lexy.Int()
	buf := codec.Append(nil, -4567890)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// -4567890
}

func ExampleFloat32() {
	codec := lexy.Float32()
	value := float32(1.45e-17)
	buf := codec.Append(nil, value)
	decoded, _ := codec.Get(buf)
	fmt.Println(math.Float32bits(value) == math.Float32bits(decoded))
	// Output:
	// true
}

func ExampleFloat64() {
	codec := lexy.Float64()
	value := math.Copysign(math.NaN(), -1.0)
	buf := codec.Append(nil, value)
	decoded, _ := codec.Get(buf)
	fmt.Println(math.Float64bits(value) == math.Float64bits(decoded))
	// Output:
	// true
}

func ExampleComplex64() {
	codec := lexy.Complex64()
	valueReal := float32(math.Inf(1))
	valueImag := float32(5.4321e-12)
	buf := codec.Append(nil, complex(valueReal, valueImag))
	decoded, _ := codec.Get(buf)
	fmt.Println(math.Float32bits(valueReal) == math.Float32bits(real(decoded)))
	fmt.Println(math.Float32bits(valueImag) == math.Float32bits(imag(decoded)))
	// Output:
	// true
	// true
}

func ExampleComplex128() {
	codec := lexy.Complex128()
	v1 := complex(123.5431, 9.87)
	v2 := complex(123.5432, 9.87)
	encodedV1 := codec.Append(nil, v1)
	encodedV2 := codec.Append(nil, v2)
	fmt.Println(bytes.Compare(encodedV1, encodedV2))
	// Output:
	// -1
}

func ExampleString() {
	codec := lexy.String()
	buf := codec.Append(nil, "")
	decoded, _ := codec.Get(buf)
	fmt.Printf("%q\n", decoded)
	buf = codec.Append(nil, "Go rocks!")
	decoded, _ = codec.Get(buf)
	fmt.Printf("%q\n", decoded)
	// Output:
	// ""
	// "Go rocks!"
}

func ExampleDuration() {
	codec := lexy.Duration()
	buf := codec.Append(nil, time.Hour*57)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// 57h0m0s
}

func ExampleTime() {
	codec := lexy.Time()
	buf := codec.Append(nil, time.Date(2000, 1, 2, 3, 4, 5, 678_901_234, time.UTC))
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded.Format(time.RFC3339Nano))
	// Output:
	// 2000-01-02T03:04:05.678901234Z
}

func ExampleBigInt() {
	codec := lexy.BigInt()
	var value big.Int
	value.SetString("-1234567890123456789012345678901234567890", 10)
	buf := codec.Append(nil, &value)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// -1234567890123456789012345678901234567890
}

func ExampleBigFloat() {
	codec := lexy.BigFloat()
	var value big.Float
	value.SetString("-1.23456789e+50732")
	buf := codec.Append(nil, &value)
	decoded, _ := codec.Get(buf)
	fmt.Println(value.Cmp(decoded))
	// Output:
	// 0
}

func ExampleBigRat() {
	codec := lexy.BigRat()
	// value will be -832/6 in lowest terms
	var value big.Rat
	var num, denom big.Int
	num.SetString("12345", 10)
	denom.SetString("-90", 10)
	value.SetFrac(&num, &denom)
	buf := codec.Append(nil, &value)
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// -823/6
}

func ExamplePointerTo() {
	codec := lexy.PointerTo(lexy.String())
	value := "abc"
	buf := codec.Append(nil, &value)
	decoded, _ := codec.Get(buf)
	fmt.Println(value == *decoded)
	fmt.Println(&value == decoded)
	// Output:
	// true
	// false
}

func ExampleSliceOf() {
	type words []string
	codec := lexy.CastSliceOf[words](lexy.String())
	buf := codec.Append(nil, words{"The", "time", "is", "now"})
	decoded, _ := codec.Get(buf)
	fmt.Printf("%T\n", decoded)
	fmt.Println(decoded)
	// Output:
	// lexy_test.words
	// [The time is now]
}

func ExampleBytes() {
	codec := lexy.Bytes()
	buf := codec.Append(nil, []byte{1, 2, 3, 11, 17})
	decoded, _ := codec.Get(buf)
	fmt.Println(decoded)
	// Output:
	// [1 2 3 11 17]
}

func ExampleMapOf() {
	type word string
	type count int
	type wordCounts map[word]count
	codec := lexy.CastMapOf[wordCounts](lexy.CastString[word](), lexy.CastInt[count]())
	value := wordCounts{
		"Now":  23,
		"is":   42,
		"the":  59,
		"time": 12,
	}
	buf := codec.Append(nil, value)
	decoded, _ := codec.Get(buf)
	fmt.Printf("%T\n", decoded)
	fmt.Printf("%T\n", decoded["not-found"])
	fmt.Println(reflect.DeepEqual(value, decoded))
	// Output:
	// lexy_test.wordCounts
	// lexy_test.count
	// true
}

func ExampleNegate() {
	// Exactly the same as the lexy.Int32() example, except negated.
	codec := lexy.Negate(lexy.Int32())
	var encoded [][]byte
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		buf := codec.Append(nil, value)
		encoded = append(encoded, buf)
	}
	// Verify the encodings are decreasing.
	for i, b := range encoded[1:] {
		fmt.Println(bytes.Compare(encoded[i], b))
	}
	// Output:
	// 1
	// 1
	// 1
	// 1
}
