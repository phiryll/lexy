package lexy

import (
	"math"
)

// Codecs used by other Codecs.
var (
	uint32Codec   Codec[uint32]  = uintCodec[uint32]{}
	uint64Codec   Codec[uint64]  = uintCodec[uint64]{}
	int8Codec     Codec[int8]    = intCodec[int8]{signBit: math.MinInt8}
	int32Codec    Codec[int32]   = intCodec[int32]{signBit: math.MinInt32}
	int64Codec    Codec[int64]   = intCodec[int64]{signBit: math.MinInt64}
	aFloat32Codec Codec[float32] = Float32Codec[float32]()
	aFloat64Codec Codec[float64] = Float64Codec[float64]()
)
