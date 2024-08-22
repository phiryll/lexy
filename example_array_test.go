package lexy_test

import (
	"bytes"
	"fmt"
	"math"

	"github.com/phiryll/lexy"
)

type Quaternion [4]float64

type quaternionCodec struct{}

var quatCodec lexy.Codec[Quaternion] = quaternionCodec{}

func (quaternionCodec) Append(buf []byte, value Quaternion) []byte {
	for i := range value {
		buf = lexy.Float64().Append(buf, value[i])
	}
	return buf
}

func (quaternionCodec) Put(buf []byte, value Quaternion) int {
	n := 0
	for i := range value {
		n += lexy.Float64().Put(buf[n:], value[i])
	}
	return n
}

func (quaternionCodec) Get(buf []byte) (Quaternion, int) {
	var value Quaternion
	if len(buf) == 0 {
		return value, -1
	}
	n := 0
	for i := range value {
		elem, size := lexy.Float64().Get(buf[n:])
		value[i] = elem
		n += size
	}
	return value, n
}

func (quaternionCodec) RequiresTerminator() bool {
	return false
}

// ExampleArray shows how to define a Codec for an array type.
// This particular example implements all Codec methods for completeness,
// but this may not be necessary depending on your use case.
// Unimplemented methods should panic.
func Example_array() {
	quats := []Quaternion{
		{0.0, 3.4, 2.1, -1.5},
		{-9.3e+10, 7.6, math.Inf(1), 42.0},
	}
	for _, quat := range quats {
		appendBuf := quatCodec.Append(nil, quat)
		putBuf := make([]byte, 4*8)
		quatCodec.Put(putBuf, quat)
		fmt.Println(bytes.Equal(appendBuf, putBuf))
		getDecoded, _ := quatCodec.Get(appendBuf)
		fmt.Println(getDecoded)
	}
	// Output:
	// true
	// [0 3.4 2.1 -1.5]
	// true
	// [-9.3e+10 7.6 +Inf 42]
}
