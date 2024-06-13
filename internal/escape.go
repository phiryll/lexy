package internal

// Functions for delimiting elements of aggregates (maps, slices, and
// structs) and escaping. These are defined in away that preserves the
// lexicographical []byte ordering of encoded aggregate types. For
// example, the encoding of ["ab", "cde"] needs to less than the
// encoding of ["abc", "de"], because "ab" is less than "abc". The
// delimiter can't itself be used to escape a delimiter because it leads
// to ambiguities, so there needs to be a distinct escape character.

// This comment explains why 0x00 and 0x01 were chosen for the delimiter
// and escape values. Strings are used in the following examples for
// clarity, with "," and "\" denoting the delimiter and escape bytes.
// The form of the examples is (input []string -> encoded string), with
// all input characters having their natural meaning (no delimiters or
// escapes). The encodings for maps and structs will be analogous.
//
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
// A, B, and C must all be less than D, E, and F. Therefore "," must be
// less than all other values, including the escape. The delimiter must
// be 0x00.
//
// Since the delimiter is less than all other values, E must be less
// than D (first slice element "a," < "ab"), so the encoded value
// "a\,,bc" must be less than "ab,c". Therefore "\" must be less than
// all other values except the delimiter. The escape must be 0x01.

// Used to delimit elements of an aggregate - slice elements, map keys
// and values, and struct field names and values.
const delimiter byte = 0x00

// Used to escape the delimiter and escape bytes when they appear in
// data. This includes appearing in the encodings of nested aggregates,
// because those are still just data at the level of the enclosing
// aggregate.
const escape byte = 0x01
