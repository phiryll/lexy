package internal

import (
	"bytes"
	"io"
)

// Helpers for encoding and decoding key-value pairs.
// Readers and writers are intentionally decoupled,
// because generic types can be inconsistent within a Codec implementation.
// Pairs are encoded as [escaped encoded key, delimeter, escaped encoded value].

type pairReader[K any, V any] struct {
	keyReader   Reader[K]
	valueReader Reader[V]
}

type pairWriter[K any, V any] struct {
	keyWriter   Writer[K]
	valueWriter Writer[V]
}

func (p pairReader[K, V]) read(r io.Reader) (K, V, error) {
	var zeroKey K
	var zeroValue V

	b, readErr := Unescape(r)
	if len(b) == 0 && readErr == io.EOF {
		return zeroKey, zeroValue, io.EOF
	}
	if readErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(readErr)
	}
	key, codecErr := p.keyReader(bytes.NewBuffer(b))
	if codecErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(codecErr)
	}

	b, readErr = Unescape(r)
	// Ignore io.EOF here.
	// valueReader should catch it if the bytes read are incomplete.
	if readErr != nil && readErr != io.EOF {
		return zeroKey, zeroValue, readErr
	}
	value, codecErr := p.valueReader(bytes.NewBuffer(b))
	if codecErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(codecErr)
	}
	return key, value, nil
}

func (p pairWriter[K, V]) write(w io.Writer, key K, value V, scratch *bytes.Buffer) error {
	scratch.Reset()
	if err := p.keyWriter(scratch, key); err != nil {
		return err
	}
	if _, err := Escape(w, scratch.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(del); err != nil {
		return err
	}

	scratch.Reset()
	if err := p.valueWriter(scratch, value); err != nil {
		return err
	}
	if _, err := Escape(w, scratch.Bytes()); err != nil {
		return err
	}
	return nil
}
