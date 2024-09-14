package lexy

import (
	"errors"
	"fmt"
)

var (
	errUnterminatedBuffer  = errors.New("no unescaped terminator found")
	errUnexpectedNilsFirst = errors.New("read nils-first prefix when nils-last was configured")
	errUnexpectedNilsLast  = errors.New("read nils-last prefix when nils-first was configured")
	errBigFloatEncoding    = errors.New("unexpected failure encoding big.Float")
)

type unknownPrefixError struct {
	prefix byte
}

func (e unknownPrefixError) Error() string {
	return fmt.Sprintf("unexpected prefix 0x%X", e.prefix)
}

type badTypeError struct {
	value any
}

func (e badTypeError) Error() string {
	return fmt.Sprintf("bad type %T", e.value)
}
