package internal_test

import (
	"bytes"
	"maps"
	"math"
	"testing"

	"github.com/phiryll/lexy/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	stringCodec := internal.StringCodec
	int32Codec := internal.Int32Codec
	codec := internal.NewMapCodec(stringCodec, int32Codec)

	// at most one key so order does not matter
	testCodec[map[string]int32](t, codec, []testCase[map[string]int32]{
		{"nil", nil, []byte(nil)},
		{"empty", map[string]int32{}, []byte{empty}},
		{"{a:0}", map[string]int32{"a": 0}, []byte{
			nonEmpty, nonEmpty, 'a', del, 0x80, esc, 0x00, esc, 0x00, esc, 0x00,
		}},
	})
	testCodecFail[map[string]int32](t, codec, map[string]int32{})

	// Can't easily test the encoded bytes, so we're testing the round trip instead.
	t.Run("non-trivial", func(t *testing.T) {
		original := map[string]int32{
			"a": 0,
			"b": -1,
			"":  1000,
			"c": math.MaxInt32,
			"d": math.MinInt32,
		}
		// shallow clone, but that's fine for string and int32
		clone := maps.Clone(original)

		var b bytes.Buffer
		err := codec.Write(&b, clone)
		require.NoError(t, err)

		got, err := codec.Read(bytes.NewReader(b.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, original, got, "round trip")
		// Just double-checking that m was not mutated
	})
}