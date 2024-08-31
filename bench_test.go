package lexy_test

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/phiryll/lexy"
)

type benchCase[T any] struct {
	name  string
	value T
}

// Types used for cast benchmarking.
type (
	MyInt32 int32
	MySlice []MyInt32
)

//nolint:revive
func BenchmarkNothing(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// do nothing
	}
}

func BenchmarkAllocate(b *testing.B) {
	for _, bb := range []benchCase[int]{
		{"0", 0},
		{"1", 1},
		{"20", 20},
		{"40", 40},
		{"60", 60},
		{"80", 80},
		{"100", 100},
		{"200", 200},
		{"400", 400},
		{"600", 600},
		{"800", 800},
		{"1000", 1000},
	} {
		bb := bb
		b.Run(bb.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]byte, bb.value)
			}
		})
	}
}

func BenchmarkEmpty(b *testing.B) {
	benchCodec(b, lexy.Empty[bool](), []benchCase[bool]{
		{"true", true},
	})
}

func BenchmarkBool(b *testing.B) {
	benchCodec(b, lexy.Bool(), []benchCase[bool]{
		{"true", true},
	})
}

func BenchmarkUint8(b *testing.B) {
	benchCodec(b, lexy.Uint8(), []benchCase[uint8]{
		{"0x7F", 0x7F},
	})
}

func BenchmarkUint16(b *testing.B) {
	benchCodec(b, lexy.Uint16(), []benchCase[uint16]{
		{"0x8000", 0x8000},
	})
}

func BenchmarkUint32(b *testing.B) {
	benchCodec(b, lexy.Uint32(), []benchCase[uint32]{
		{"0x00000001", 0x00000001},
	})
}

func BenchmarkUint64(b *testing.B) {
	benchCodec(b, lexy.Uint64(), []benchCase[uint64]{
		{"0xFFFFFFFFFFFFFFFF", 0xFFFFFFFFFFFFFFFF},
	})
}

func BenchmarkUint(b *testing.B) {
	benchCodec(b, lexy.Uint(), []benchCase[uint]{
		{"1", 1},
	})
}

func BenchmarkInt8(b *testing.B) {
	benchCodec(b, lexy.Int8(), []benchCase[int8]{
		{"min", math.MinInt8},
	})
}

func BenchmarkInt16(b *testing.B) {
	benchCodec(b, lexy.Int16(), []benchCase[int16]{
		{"-1", -1},
	})
}

func BenchmarkInt32(b *testing.B) {
	benchCodec(b, lexy.Int32(), []benchCase[int32]{
		{"+1", 1},
	})
}

func BenchmarkCastInt32(b *testing.B) {
	benchCodec(b, lexy.CastInt32[MyInt32](), []benchCase[MyInt32]{
		{"+1", 1},
	})
}

func BenchmarkInt64(b *testing.B) {
	benchCodec(b, lexy.Int64(), []benchCase[int64]{
		{"max", math.MaxInt64},
	})
}

func BenchmarkInt(b *testing.B) {
	benchCodec(b, lexy.Int(), []benchCase[int]{
		{"-1", -1},
	})
}

func BenchmarkFloat32(b *testing.B) {
	benchCodec(b, lexy.Float32(), []benchCase[float32]{
		{"4.1298e-18", 4.1298e-18},
	})
}

func BenchmarkFloat64(b *testing.B) {
	benchCodec(b, lexy.Float64(), []benchCase[float64]{
		{"-7.123874e+24", -7.123874e+24},
	})
}

func BenchmarkComplex64(b *testing.B) {
	benchCodec(b, lexy.Complex64(), []benchCase[complex64]{
		{"(-Inf, 2.6431329)", complex(float32(math.Inf(-1)), float32(2.6431329))},
	})
}

func BenchmarkComplex128(b *testing.B) {
	benchCodec(b, lexy.Complex128(), []benchCase[complex128]{
		{"(324.148, NaN)", complex(float64(324.148), math.NaN())},
	})
}

func BenchmarkString(b *testing.B) {
	benchCodec(b, lexy.String(), []benchCase[string]{
		{"empty", ""},
		{"1 byte", "a"},
		{"20 bytes", "12345678901234567890"},
		{"1000 bytes", string(randomBytes(1000, 44327819))},
	})
}

func BenchmarkTime(b *testing.B) {
	locLA, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic("could not get time zone")
	}
	date1900 := time.Date(1900, 1, 2, 3, 4, 5, 6, time.UTC)
	date2000 := time.Date(2000, 1, 2, 3, 4, 5, 6, locLA)
	// Prime the timezone format cache.
	codec := lexy.Time()
	codec.Get(codec.Append(nil, date1900))
	codec.Get(codec.Append(nil, date2000))
	benchCodec(b, codec, []benchCase[time.Time]{
		{"1900 UTC", date1900},
		{"2000 LA", date2000},
	})
}

func BenchmarkDuration(b *testing.B) {
	benchCodec(b, lexy.Duration(), []benchCase[time.Duration]{
		{"1000000 ns", 1000000 * time.Nanosecond},
	})
}

func BenchmarkBigInt(b *testing.B) {
	benchCodec(b, lexy.BigInt(), []benchCase[*big.Int]{
		{"0", big.NewInt(0)},
		{"-1", big.NewInt(-1)},
		{"+1", big.NewInt(1)},
		{"big pos", newBigInt(manyDigits)},
		{"big neg", newBigInt("-" + manyDigits)},
	})
}

func BenchmarkBigFloat(b *testing.B) {
	var negZero, posZero, negInf, posInf big.Float
	negZero.Neg(&negZero)
	negInf.SetInf(true)
	posInf.SetInf(false)
	benchCodec(b, lexy.BigFloat(), []benchCase[*big.Float]{
		{"-0", &negZero},
		{"+0", &posZero},
		{"-Inf", &negInf},
		{"+Inf", &posInf},
		{"big pos whole", newBigFloat(manyDigits + manyDigits)},
		{"big pos mixed", newBigFloat(manyDigits + "." + manyDigits)},
		{"big pos small", newBigFloat("0." + manyZeros + manyDigits)},
		{"big neg whole", newBigFloat("-" + manyDigits + manyDigits)},
		{"big neg mixed", newBigFloat("-" + manyDigits + "." + manyDigits)},
		{"big neg small", newBigFloat("-0." + manyZeros + manyDigits)},
	})
}

func BenchmarkBigRat(b *testing.B) {
	benchCodec(b, lexy.BigRat(), []benchCase[*big.Rat]{
		{"-1/3", newBigRat("-1", "3")},
		{"0/123", newBigRat("0", "123")},
		{"pos big/big", newBigRat(manyDigits, "42"+manyDigits)},
		{"neg big/big", newBigRat("-"+manyDigits, "42"+manyDigits)},
	})
}

func BenchmarkBytes(b *testing.B) {
	benchCodec(b, lexy.Bytes(), []benchCase[[]byte]{
		{"nil", nil},
		{"empty", []byte{}},
		{"1 byte", []byte{53}},
		{"20 bytes", []byte("12345678901234567890")},
		{"1000 bytes", randomBytes(1000, 3891217)},
	})
}

// Keeping aggregates relatively simple,
// because we're benchmarking the aggregation, not the elements.

func BenchmarkPointerTo(b *testing.B) {
	benchCodec(b, lexy.PointerTo(lexy.Int32()), []benchCase[*int32]{
		{"nil", nil},
		{"*0", ptr(int32(0))},
		{"*-1", ptr(int32(-1))},
	})
}

func BenchmarkSliceOf(b *testing.B) {
	benchCodec(b, lexy.SliceOf(lexy.Int32()), []benchCase[[]int32]{
		{"nil", nil},
		{"empty", []int32{}},
		{"1 element", []int32{931}},
		{"1000 elements", randomInt32(1000, 28931)},
	})
}

func BenchmarkCastSliceOf(b *testing.B) {
	slice := randomInt32(1000, 28931)
	bigSlice := make(MySlice, 1000)
	for i := 0; i < 1000; i++ {
		bigSlice[i] = MyInt32(slice[i])
	}
	benchCodec(b, lexy.CastSliceOf[MySlice](lexy.CastInt32[MyInt32]()), []benchCase[MySlice]{
		{"nil", nil},
		{"empty", MySlice{}},
		{"1 element", MySlice{931}},
		{"1000 elements", bigSlice},
	})
}

// Timing how long it takes to build maps used in BenchmarkMapOf,
// to separate Get() performance from just building the map.
func BenchmarkRawMap(b *testing.B) {
	ints := randomInt32(2000, 639871)
	for _, bb := range []benchCase[[]int32]{
		{"empty", []int32{}},
		{"1 element", []int32{43943, -319432}},
		{"1000 elements", ints},
	} {
		bb := bb
		b.Run(bb.name, func(b *testing.B) {
			arr := bb.value
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := map[int32]int32{}
				for k := 0; k < len(arr); k += 2 {
					m[arr[k]] = m[arr[k+1]]
				}
			}
		})
	}
}

func BenchmarkMapOf(b *testing.B) {
	ints := randomInt32(2000, 639871)
	bigMap := make(map[int32]int32, 1000)
	for i := 0; i < 1000; i++ {
		bigMap[ints[2*i]] = ints[2*i+1]
	}
	benchCodec(b, lexy.MapOf(lexy.Int32(), lexy.Int32()), []benchCase[map[int32]int32]{
		{"nil", nil},
		{"empty", map[int32]int32{}},
		{"1 element", map[int32]int32{43943: -319432}},
		{"1000 elements", bigMap},
	})
}

func BenchmarkNegate(b *testing.B) {
	benchCodec(b, lexy.Negate(lexy.BigInt()), []benchCase[*big.Int]{
		{"0", big.NewInt(0)},
		{"-1", big.NewInt(-1)},
		{"+1", big.NewInt(1)},
		{"big pos", newBigInt(manyDigits)},
		{"big neg", newBigInt("-" + manyDigits)},
	})
}

func BenchmarkNegateEscaped(b *testing.B) {
	benchCodec(b, lexy.Negate(lexy.Bytes()), []benchCase[[]byte]{
		{"nil", nil},
		{"empty", []byte{}},
		{"1 byte", []byte{53}},
		{"20 bytes", randomBytes(20, 601239)},
		{"40 bytes", randomBytes(40, 9312457)},
		{"60 bytes", randomBytes(60, 38701)},
		{"80 bytes", randomBytes(80, 5239107)},
		{"100 bytes", randomBytes(100, 4387201)},
		{"200 bytes", randomBytes(200, 23832)},
		{"400 bytes", randomBytes(400, 129045)},
		{"600 bytes", randomBytes(600, 7932462)},
		{"800 bytes", randomBytes(800, 931247)},
		{"1000 bytes", randomBytes(1000, 3903748)},
	})
}

func BenchmarkTerminate(b *testing.B) {
	benchCodec(b, lexy.Terminate(lexy.Bytes()), []benchCase[[]byte]{
		{"nil", nil},
		{"empty", []byte{}},
		{"1 byte", []byte{53}},
		{"20 bytes", randomBytes(20, 601239)},
		{"40 bytes", randomBytes(40, 9312457)},
		{"60 bytes", randomBytes(60, 38701)},
		{"80 bytes", randomBytes(80, 5239107)},
		{"100 bytes", randomBytes(100, 4387201)},
		{"200 bytes", randomBytes(200, 23832)},
		{"400 bytes", randomBytes(400, 129045)},
		{"600 bytes", randomBytes(600, 7932462)},
		{"800 bytes", randomBytes(800, 931247)},
		{"1000 bytes", randomBytes(1000, 3903748)},
	})
}

//nolint:gosec,revive
func randomBytes(n int, seed int64) []byte {
	random := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	random.Read(b)
	return b
}

//nolint:gosec
func randomInt32(n int, seed int64) []int32 {
	random := rand.New(rand.NewSource(seed))
	b := make([]int32, n)
	for i := 0; i < n; i++ {
		b[i] = int32(random.Uint32())
	}
	return b
}

//nolint:thelper
func benchCodec[T any](b *testing.B, codec lexy.Codec[T], benchCases []benchCase[T]) {
	if len(benchCases) == 1 {
		benchSingleValue(b, codec, benchCases[0].value)
		return
	}
	for _, bb := range benchCases {
		bb := bb
		b.Run(bb.name, func(b *testing.B) {
			benchSingleValue(b, codec, bb.value)
		})
	}
}

//nolint:thelper
func benchSingleValue[T any](b *testing.B, codec lexy.Codec[T], value T) {
	// Tests both encoding and how efficiently codec.Append allocates the buffer.
	b.Run("append nil", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			codec.Append(nil, value)
		}
	})
	// Tests just encoding, because the buffer is already big enough.
	// This should be very close in performance to codec.Put.
	b.Run("append", func(b *testing.B) {
		buf := codec.Append(nil, value)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			codec.Append(buf[:0], value)
		}
	})
	b.Run("put", func(b *testing.B) {
		buf := codec.Append(nil, value)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			codec.Put(buf, value)
		}
	})
	b.Run("get", func(b *testing.B) {
		buf := codec.Append(nil, value)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			codec.Get(buf)
		}
	})
}
