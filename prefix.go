package lexy

// Prefixes to use for encodings for types whose instances can be nil.
// The values were chosen so that nils-first < non-nil < nils-last,
// and neither the prefixes nor their complements need to be escaped.
const (
	// Room for more between non-nil and nils-last if needed.
	prefixNilFirst byte = 0x02
	prefixNonNil   byte = 0x03
	prefixNilLast  byte = 0xFD
)

// Convenience byte slices.
var (
	pNilFirst = []byte{prefixNilFirst}
	pNonNil   = []byte{prefixNonNil}
	pNilLast  = []byte{prefixNilLast}
)
