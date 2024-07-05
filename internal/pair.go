package internal

import (
	"bytes"
	"io"
)

// Helper type for encoding and decoding key-value pairs.
// This does not implement Codec, because it does not encode a single value.
// Creating a new Pair type is overkill for the map and struct use cases.
// read() will only return io.EOF if the underying Reader does and zero bytes were read.
//
// Pairs are encoded as [escaped encoded key, delimeter, escaped encoded value].
type pairCodec[K any, V any] struct {
	keyCodec   codec[K]
	valueCodec codec[V]
}

func newPairCodec[K any, V any](keyCodec codec[K], valueCodec codec[V]) pairCodec[K, V] {
	// TODO: use default if possible based on types
	if keyCodec == nil {
		panic("keyCodec must be non-nil")
	}
	if valueCodec == nil {
		panic("valueCodec must be non-nil")
	}
	return pairCodec[K, V]{keyCodec, valueCodec}
}

func (c pairCodec[K, V]) read(r io.Reader) (K, V, error) {
	var zeroKey K
	var zeroValue V

	b, readErr := Unescape(r)
	if len(b) == 0 && readErr == io.EOF {
		return zeroKey, zeroValue, io.EOF
	}
	if readErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(readErr)
	}
	key, codecErr := c.keyCodec.Read(bytes.NewBuffer(b))
	if codecErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(codecErr)
	}

	b, readErr = Unescape(r)
	// Ignore io.EOF here.
	// valueCodec.Read should catch it if the bytes read are incomplete.
	if readErr != nil && readErr != io.EOF {
		return zeroKey, zeroValue, readErr
	}
	value, codecErr := c.valueCodec.Read(bytes.NewBuffer(b))
	if codecErr != nil {
		return zeroKey, zeroValue, unexpectedIfEOF(codecErr)
	}
	return key, value, nil
}

func (c pairCodec[K, V]) write(w io.Writer, key K, value V, scratch *bytes.Buffer) error {
	scratch.Reset()
	if err := c.keyCodec.Write(scratch, key); err != nil {
		return err
	}
	if _, err := Escape(w, scratch.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(del); err != nil {
		return err
	}

	scratch.Reset()
	if err := c.valueCodec.Write(scratch, value); err != nil {
		return err
	}
	if _, err := Escape(w, scratch.Bytes()); err != nil {
		return err
	}
	return nil
}
