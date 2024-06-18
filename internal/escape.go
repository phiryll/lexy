package internal

// Functions for delimiting elements of aggregates (maps, slices, and structs) and escaping.
// The lexicographical binary ordering of encoded aggregates is preserved.
// For example, ["ab", "cde"] is less than ["aba", "de"], because "ab" is less than "aba".
// The delimiter can't itself be used to escape a delimiter because it leads to ambiguities,
// so there needs to be a distinct escape character.

// This comment explains why the delimiter and escape values must be 0x00 and 0x01.
// Strings are used for clarity, with "," and "\" denoting the delimiter and escape bytes.
// All input characters have their natural meaning (no delimiters or escapes).
// The encodings for maps and structs will be analogous.
//
//      input slice  -> encoded string
//   A: ["a", "bc"]  -> a,bc
//   B: ["a", ",bc"] -> a,\,bc
//   C: ["a", "\bc"] -> a,\\bc
//   D: ["ab", "c"]  -> ab,c
//   E: ["a,", "bc"] -> a\,,bc
//   F: ["a\", "bc"] -> a\\,bc
//
// B and E are an example of why the delimiter can't be its own escape,
// the encoded strings would both be "a,,,b".
//
// A, B, and C must all be less than D, E, and F.
// We can see "," must be less than all other values including the escape, so it must be 0x00.
//
// Since "," is less than everything else, E < D (first slice element "a," < "ab"). Therefore "a\,,bc" < "ab,c".
// We can see "/" must be less than all other values except the delimiter, so it must be 0x01.

// delimiter is used to delimit elements of an aggregate value.
const delimiter byte = 0x00

// escape is used the escape the delimiter and escape bytes when they appear in data.
//
// This includes appearing in the encodings of nested aggregates,
// because those are still just data at the level of the enclosing aggregate.
const escape byte = 0x01
