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

var (
	TestingDoEscape   = doEscape
	TestingDoUnescape = doUnescape
)
