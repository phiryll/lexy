package internal_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
)

func TestNegateInt32(t *testing.T) {
	codec := internal.MakeNegateCodec(int32Codec)
	testCodecRoundTrip(t, codec, []testCase[int32]{
		{"min", math.MinInt32, nil},
		{"-1", -1, nil},
		{"0", 0, nil},
		{"+1", 1, nil},
		{"max", math.MaxInt32, nil},
	})

	encode := func(value int32) []byte {
		var b bytes.Buffer
		if err := codec.Write(&b, value); err != nil {
			panic(err)
		}
		return b.Bytes()
	}
	assert.IsIncreasing(t, [][]byte{
		encode(math.MaxInt32),
		encode(100),
		encode(1),
		encode(0),
		encode(-1),
		encode(-100),
		encode(math.MinInt32),
	})
}

func TestNegatePtrString(t *testing.T) {
	ptrCodec := internal.MakePointerCodec[*string](stringCodec)
	codec := internal.MakeNegateCodec(ptrCodec)
	testCodecRoundTrip(t, codec, []testCase[*string]{
		{"nil", nil, nil},
		{"*empty", ptr(""), nil},
		{"*abc", ptr("abc"), nil},
		{"*def", ptr("def"), nil},
	})

	encode := func(value *string) []byte {
		var b bytes.Buffer
		if err := codec.Write(&b, value); err != nil {
			panic(err)
		}
		return b.Bytes()
	}
	assert.IsIncreasing(t, [][]byte{
		encode(ptr("def")),
		encode(ptr("abc")),
		encode(ptr("")),
		// encode(nil),
	})
}
