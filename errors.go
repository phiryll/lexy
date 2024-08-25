package lexy

import (
	"errors"
	"fmt"
)

var (
	errNil                 = errors.New("cannot be nil")
	errUnexpectedNilsFirst = errors.New("read nils-first prefix when nils-last was configured")
	errUnexpectedNilsLast  = errors.New("read nils-last prefix when nils-first was configured")
	errBigFloatEncoding    = errors.New("unexpected failure encoding big.Float")
)

type unknownPrefixError struct {
	prefix byte
}

func (e unknownPrefixError) Error() string {
	return fmt.Sprintf("unexpected prefix %X", e.prefix)
}

type badTypeError struct {
	value any
}

func (e badTypeError) Error() string {
	return fmt.Sprintf("bad type %T", e.value)
}
