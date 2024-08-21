package lexy

import (
	"errors"
	"fmt"
)

var (
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

type nilError struct {
	name string
}

func (e nilError) Error() string {
	return e.name + " must be non-nil"
}
