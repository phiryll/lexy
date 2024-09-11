package lexy

// Things that need to be exported for testing, but should not be part of the public API.
// The identifiers are in the lexy package, but the filename ends in _test.go,
// preventing their inclusion in the public API.

const (
	TestingPrefixNilFirst = prefixNilFirst
	TestingPrefixNonNil   = prefixNonNil
	TestingPrefixNilLast  = prefixNilLast

	TestingTerminator = terminator
	TestingEscape     = escape
)

// Used by fuzz testers.
var (
	TestingTermUint16  Codec[uint16]  = terminatorCodec[uint16]{Uint16()}
	TestingTermUint64  Codec[uint64]  = terminatorCodec[uint64]{Uint64()}
	TestingTermInt16   Codec[int16]   = terminatorCodec[int16]{Int16()}
	TestingTermInt64   Codec[int64]   = terminatorCodec[int64]{Int64()}
	TestingTermFloat32 Codec[float32] = terminatorCodec[float32]{Float32()}
	TestingTermFloat64 Codec[float64] = terminatorCodec[float64]{Float64()}
)
