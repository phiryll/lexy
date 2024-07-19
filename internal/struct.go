package internal

// TODO: Add helpers for making custom struct codecs.

// Examples must cover:
//
// every supported type, and a very complex example
//
// if fields never change vs. can change
// this is likely not as tolerant as JSON serialization,
// but could be if the effort is made
// - field added
// - field removed
// - field renamed
// - field type changed
// - sort order changed
