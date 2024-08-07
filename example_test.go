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
	codec := lexy.MapOf[set](lexy.Uint8[uint8](), lexy.Empty[present]())
	var buf bytes.Buffer
	value := set{
		23: present{},
		42: present{},
		59: present{},
		12: present{},
	}
	if err := codec.Write(&buf, value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T\n", decoded)
	fmt.Printf("%T\n", decoded[0])
	fmt.Println(reflect.DeepEqual(value, decoded))
	// Output:
	// lexy_test.set
	// lexy_test.present
	// true
}

func ExampleBool() {
	codec := lexy.Bool[bool]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, true); err != nil {
		panic(err)
	}
	first, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	buf.Reset()
	if err := codec.Write(&buf, false); err != nil {
		panic(err)
	}
	second, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%t, %t", first, second)
	// Output:
	// true, false
}

func ExampleUint64() {
	codec := lexy.Uint64[uint64]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, 123); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output:
	// 123
}

func ExampleUint() {
	codec := lexy.Uint[uint]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, 4567890); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output:
	// 4567890
}

func ExampleUint32_underlyingType() {
	type size uint32
	codec := lexy.Uint32[size]()
	var buf bytes.Buffer
	// Go will type a constant appropriately.
	if err := codec.Write(&buf, 123); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Value %d of type %T", value, value)
	// Output:
	// Value 123 of type lexy_test.size
}

func ExampleInt32() {
	codec := lexy.Int32[int32]()
	var buf bytes.Buffer
	var encoded [][]byte
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		buf.Reset()
		if err := codec.Write(&buf, value); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
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
	codec := lexy.Int[int]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, -4567890); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output:
	// -4567890
}

func ExampleFloat32() {
	codec := lexy.Float32[float32]()
	value := float32(1.45e-17)
	var buf bytes.Buffer
	if err := codec.Write(&buf, value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(math.Float32bits(value) == math.Float32bits(decoded))
	// Output:
	// true
}

func ExampleFloat64() {
	codec := lexy.Float64[float64]()
	value := math.Copysign(math.NaN(), -1.0)
	var buf bytes.Buffer
	if err := codec.Write(&buf, value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(math.Float64bits(value) == math.Float64bits(decoded))
	// Output:
	// true
}

func ExampleComplex64() {
	codec := lexy.Complex64()
	valueReal := float32(math.Inf(1))
	valueImag := float32(5.4321e-12)
	var buf bytes.Buffer
	if err := codec.Write(&buf, complex(valueReal, valueImag)); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
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
	var buf bytes.Buffer
	if err := codec.Write(&buf, v1); err != nil {
		panic(err)
	}
	encodedV1 := bytes.Clone(buf.Bytes())
	buf.Reset()
	if err := codec.Write(&buf, v2); err != nil {
		panic(err)
	}
	encodedV2 := bytes.Clone(buf.Bytes())
	fmt.Println(bytes.Compare(encodedV1, encodedV2))
	// Output:
	// -1
}

func ExampleString() {
	codec := lexy.String[string]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, ""); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", decoded)
	buf.Reset()
	if err := codec.Write(&buf, "Go rocks!"); err != nil {
		panic(err)
	}
	decoded, err = codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", decoded)
	// Output:
	// ""
	// "Go rocks!"
}

func ExampleDuration() {
	codec := lexy.Duration()
	var buf bytes.Buffer
	duration := time.Hour * 57
	if err := codec.Write(&buf, duration); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output:
	// 57h0m0s
}

func ExampleTime() {
	codec := lexy.Time()
	var buf bytes.Buffer
	aTime := time.Date(2000, 1, 2, 3, 4, 5, 678_901_234, time.UTC)
	if err := codec.Write(&buf, aTime); err != nil {
		panic(err)
	}
	value, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value.Format(time.RFC3339Nano))
	// Output:
	// 2000-01-02T03:04:05.678901234Z
}

func ExampleBigInt() {
	codec := lexy.BigInt()
	var buf bytes.Buffer
	var value big.Int
	value.SetString("-1234567890123456789012345678901234567890", 10)
	if err := codec.Write(&buf, &value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(decoded)
	// Output:
	// -1234567890123456789012345678901234567890
}

func ExampleBigFloat() {
	codec := lexy.BigFloat()
	var buf bytes.Buffer
	var value big.Float
	value.SetString("-1.23456789e+50732")
	if err := codec.Write(&buf, &value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value.Cmp(decoded))
	// Output:
	// 0
}

func ExampleBigRat() {
	codec := lexy.BigRat()
	var buf bytes.Buffer
	var value big.Rat
	// value will be -832/6 in lowest terms
	var num, denom big.Int
	num.SetString("12345", 10)
	denom.SetString("-90", 10)
	value.SetFrac(&num, &denom)
	if err := codec.Write(&buf, &value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(decoded)
	// Output:
	// -823/6
}

func ExamplePointerTo() {
	codec := lexy.PointerTo[*string](lexy.String[string]())
	value := "abc"
	var buf bytes.Buffer
	if err := codec.Write(&buf, &value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(value == *decoded)
	fmt.Println(&value == decoded)
	// Output:
	// true
	// false
}

func ExampleSliceOf() {
	type words []string
	codec := lexy.SliceOf[words](lexy.String[string]())
	var buf bytes.Buffer
	if err := codec.Write(&buf, words{"The", "time", "is", "now"}); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T\n", decoded)
	fmt.Println(decoded)
	// Output:
	// lexy_test.words
	// [The time is now]
}

func ExampleBytes() {
	codec := lexy.Bytes[[]byte]()
	var buf bytes.Buffer
	if err := codec.Write(&buf, []byte{1, 2, 3, 11, 17}); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(decoded)
	// Output:
	// [1 2 3 11 17]
}

func ExampleMapOf() {
	type word string
	type count int
	type wordCounts map[word]count
	codec := lexy.MapOf[wordCounts](lexy.String[word](), lexy.Int[count]())
	var buf bytes.Buffer
	value := wordCounts{
		"Now":  23,
		"is":   42,
		"the":  59,
		"time": 12,
	}
	if err := codec.Write(&buf, value); err != nil {
		panic(err)
	}
	decoded, err := codec.Read(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T\n", decoded)
	fmt.Printf("%T\n", decoded["not-found"])
	fmt.Println(reflect.DeepEqual(value, decoded))
	// Output:
	// lexy_test.wordCounts
	// lexy_test.count
	// true
}

func ExampleNegate() {
	// Exactly the same as the lexy.Int() example, except negated.
	codec := lexy.Negate(lexy.Int32[int32]())
	var buf bytes.Buffer
	var encoded [][]byte
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		buf.Reset()
		if err := codec.Write(&buf, value); err != nil {
			panic(err)
		}
		encoded = append(encoded, bytes.Clone(buf.Bytes()))
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

func ExampleEncode() {
	codec := lexy.Int32[int32]()
	var buf bytes.Buffer
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		encoded, err := lexy.Encode(codec, value)
		if err != nil {
			panic(err)
		}
		buf.Reset()
		if err := codec.Write(&buf, value); err != nil {
			panic(err)
		}
		fmt.Println(bytes.Equal(encoded, buf.Bytes()))
	}
	// Output:
	// true
	// true
	// true
	// true
	// true
}

func ExampleDecode() {
	codec := lexy.Int32[int32]()
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
	} {
		encoded, err := lexy.Encode(codec, value)
		if err != nil {
			panic(err)
		}
		decoded, err := lexy.Decode(codec, encoded)
		if err != nil {
			panic(err)
		}
		fmt.Println(decoded)
	}
	// Output:
	// -2147483648
	// -1
	// 0
	// 1
	// 2147483647
}
