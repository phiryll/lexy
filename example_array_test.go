package lexy_test

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/phiryll/lexy"
)

type Quaternion [4]float64

var elemCodec = lexy.Float64[float64]()

type quaternionCodec struct{}

func (c quaternionCodec) Read(r io.Reader) (Quaternion, error) {
	var zero, value Quaternion
	for i := range value {
		elem, err := elemCodec.Read(r)
		if err != nil {
			if i == 0 {
				return zero, err
			}
			return zero, lexy.UnexpectedIfEOF(err)
		}
		value[i] = elem
	}
	return value, nil
}

func (c quaternionCodec) Write(w io.Writer, value Quaternion) error {
	for i := range value {
		if err := elemCodec.Write(w, value[i]); err != nil {
			return err
		}
	}
	return nil
}

func (c quaternionCodec) RequiresTerminator() bool {
	return false
}

func Example_array() {
	codec := quaternionCodec{}
	quats := []Quaternion{
		{0.0, 3.4, 2.1, -1.5},
		{math.NaN(), 7.6, math.Inf(1), 42.0},
	}
	for _, quat := range quats {
		var buf bytes.Buffer
		if err := codec.Write(&buf, quat); err != nil {
			panic(err)
		}
		decoded, err := codec.Read(&buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(decoded)
	}
	// Output:
	// [0 3.4 2.1 -1.5]
	// [NaN 7.6 +Inf 42]
}
