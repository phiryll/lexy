package lexy_test

// This file contains things that help in writing Codec tests,
// it doesn't have any tests itself.

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/phiryll/lexy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](value T) *T {
	return &value
}

func encoderFor[T any](codec lexy.Codec[T]) func(value T) []byte {
	return func(value T) []byte {
		return codec.Append(nil, value)
	}
}

func errorForEOF(bytesRead int) error {
	if bytesRead == 0 {
		return io.EOF
	}
	return io.ErrUnexpectedEOF
}

func toCodec[T any](codec lexy.NillableCodec[T]) lexy.Codec[T] {
	return codec
}

func concat(slices ...[]byte) []byte {
	var result []byte
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func makeBigBuf(size int) []byte {
	buf := make([]byte, size+100)
	for i := range buf {
		buf[i] = 37
	}
	return buf
}

func checkBigBuf(t *testing.T, buf []byte, size int) {
	for i := range buf[size:] {
		k := size + i
		assert.Equal(t, byte(37), buf[k], "buf[%d] = %d written to buffer", k, buf[k])
	}
}

type testCase[T any] struct {
	name  string
	value T
	data  []byte
}

// Returns new test cases with each testCase.data set to codec.Append(nil, testCase.value).
// This is useful when the encoded value is difficult to calculate by hand.
func fillTestData[T any](codec lexy.Codec[T], tests []testCase[T]) []testCase[T] {
	newTests := make([]testCase[T], len(tests))
	for i, tt := range tests {
		test := tt
		test.data = codec.Append(nil, tt.value)
		newTests[i] = test
	}
	return newTests
}

// The same as iotest.TruncateWriter, except it's not silent.
type boundedWriter struct {
	w io.Writer
	n int
}

func (t *boundedWriter) Write(p []byte) (int, error) {
	if t.n <= 0 {
		return 0, errWrite
	}
	// real write
	n := len(p)
	var over bool
	if n > t.n {
		n = int(t.n)
		over = true
	}
	n, err := t.w.Write(p[0:n])
	t.n -= n
	if err == nil {
		n = len(p)
	}
	if over && err != nil {
		return n, errWrite
	}
	return n, err
}

// Just to make the test cases terser.
const (
	term      byte = lexy.TestingTerminator
	esc       byte = lexy.TestingEscape
	pNilFirst byte = lexy.TestingPrefixNilFirst
	pNonNil   byte = lexy.TestingPrefixNonNil
	pNilLast  byte = lexy.TestingPrefixNilLast
)

//nolint:thelper
func testCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, false}.test(t, tests)
}

//nolint:thelper
func testVaryingCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	testerCodec[T]{codec, true}.test(t, tests)
}

// A testing wrapper for Codec that deals only in []buf at the API level.
type testerCodec[T any] struct {
	codec   lexy.Codec[T]
	varying bool
}

// The output of one Append/Put/Write method.
type output struct {
	name string
	buf  []byte
}

//nolint:thelper
func (c testerCodec[T]) test(t *testing.T, tests []testCase[T]) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// TODO: check to see that Get/Read did not change the buffer they were given.
			var outputs []output
			// Test Append/Put/Write.
			outputs = append(outputs, c.testAppendNil(t, tt))
			outputs = append(outputs, c.testAppendExisting(t, tt))
			outputs = append(outputs, c.testPut(t, tt))
			outputs = append(outputs, c.testPutLongBuf(t, tt))
			outputs = append(outputs, c.testWrite(t, tt))

			// Test Get/Read
			// referenceBuf := c.codec.Append(nil, tt.value)
			//
			// if not varying, already tested buf == tt.data
			// because it's nice to be in the same t.Run(name)

			// Test buffers that are too short by 1 byte.
			c.testPutShortBuf(t, tt)
			c.testWriteShortBuf(t, tt)
			c.testWriteTruncatedBuf(t, tt)
		})
	}
}

//nolint:thelper
func (c testerCodec[T]) testAppendNil(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("append nil", func(t *testing.T) {
		t.Parallel()
		buf = c.codec.Append(nil, tt.value)
		if buf == nil {
			buf = []byte{}
		}
		if !c.varying {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"append nil", buf}
}

//nolint:thelper
func (c testerCodec[T]) testAppendExisting(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("append existing", func(t *testing.T) {
		t.Parallel()
		header := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buf = append(buf, header...)
		buf = c.codec.Append(buf, tt.value)
		assert.Equal(t, header, buf[:len(header)])
		buf = buf[len(header):]
		if !c.varying {
			assert.Equal(t, tt.data, buf)
		} else {
			size := len(c.codec.Append(nil, tt.value))
			assert.Equal(t, size, len(buf))
		}
	})
	return output{"append existing", buf}
}

//nolint:thelper
func (c testerCodec[T]) testPut(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put", func(t *testing.T) {
		t.Parallel()
		size := len(c.codec.Append(nil, tt.value))
		buf = make([]byte, size)
		putSize := c.codec.Put(buf, tt.value)
		assert.Equal(t, size, putSize)
		if !c.varying {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put", buf}
}

//nolint:thelper
func (c testerCodec[T]) testPutLongBuf(t *testing.T, tt testCase[T]) output {
	var buf []byte
	t.Run("put long buf", func(t *testing.T) {
		t.Parallel()
		size := len(c.codec.Append(nil, tt.value))
		buf = makeBigBuf(size)
		putSize := c.codec.Put(buf, tt.value)
		assert.Equal(t, size, putSize)
		checkBigBuf(t, buf, size)
		buf = buf[:putSize]
		if !c.varying {
			assert.Equal(t, tt.data, buf)
		}
	})
	return output{"put long buf", buf}
}

//nolint:thelper
func (c testerCodec[T]) testPutShortBuf(t *testing.T, tt testCase[T]) {
	t.Run("put short buf", func(t *testing.T) {
		t.Parallel()
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := makeBigBuf(size)
		assert.Panics(t, func() {
			c.codec.Put(buf[:size-1], tt.value)
		})
	})
}

//nolint:thelper
func (c testerCodec[T]) testWrite(t *testing.T, tt testCase[T]) output {
	buf := bytes.NewBuffer([]byte{})
	t.Run("write", func(t *testing.T) {
		t.Parallel()
		err := c.codec.Write(buf, tt.value)
		require.NoError(t, err)
		if !c.varying {
			assert.Equal(t, tt.data, buf.Bytes())
		} else {
			size := len(c.codec.Append(nil, tt.value))
			assert.Equal(t, size, buf.Len())
		}
	})
	return output{"write", buf.Bytes()}
}

//nolint:thelper
func (c testerCodec[T]) testWriteShortBuf(t *testing.T, tt testCase[T]) {
	t.Run("write short buf", func(t *testing.T) {
		t.Parallel()
		t.Skip("Write does not yet fail properly on noisy truncation.")
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := bytes.NewBuffer([]byte{})
		w := boundedWriter{buf, size - 1}
		err := c.codec.Write(&w, tt.value)
		require.Error(t, err)
	})
}

//nolint:thelper
func (c testerCodec[T]) testWriteTruncatedBuf(t *testing.T, tt testCase[T]) {
	t.Run("write truncated buf", func(t *testing.T) {
		t.Parallel()
		t.Skip("Write does not yet fail properly on silent truncation.")
		size := len(c.codec.Append(nil, tt.value))
		if size == 0 {
			return
		}
		buf := bytes.NewBuffer([]byte{})
		err := c.codec.Write(iotest.TruncateWriter(buf, int64(size-1)), tt.value)
		require.Error(t, err)
	})
}

/*
	Top level testing functions, need to name these differently from others!

	testCodec( testing.T, codec, []testCase )
	  t.Run("test codec")
		append/put/get/write/read( testing.T, codec, []testCase )
		  t.Run("append nil")
		    for range tests { t.Run(tt.name) ... }
		  t.Run("append existing")
		  ...

	testCodecMakeData( testing.T, codec, []testCase )
	  newTest := []testCase with codec.Append(nil, tt.value)
	  testCodec(t, codec, newTests)

	Tests are VERY similar, can we figure out a better way
	to share this code? Maybe bufCodec bool to say which tests to do?
	testVaryingCodec( testing.T, codec, []testCase )
	  bufCodec := bufferCodec{codec}
	  t.Run("round trip")
		for range tests
		  t.Run(tt.name)
			for buf := range bufCodec.AppendNil/AppendExisting/Put/Write(t, tt.value)
			  testRoundTripBuf(t, bufCodec, tt.value, buf.Name, buf)
				t.Run(buf.Name+"-get")
				  test bufCodec.Get(buf) == tt.value, and type
				t.Run(buf.Name+"-read")
				  test bufCodec.Read(buf) == tt.value, and type
	  t.Run("short buff")
		for range tests
		  t.Run(tt.name+"-put")
			bufCodec.PutShortBuf(t, tt.value)
			  panics w/ -1 byte
		  t.Run(tt.name+"-get")
			bufCodec.GetShortBuf(t, tt.value)
			  panics w/ -1 byte // maybe just return a custom error instead?
			  OR got size-1 and wrong value
		  t.Run(tt.name+"-read")
			bufCodec.ReadShortBuf(t, tt.value)
			  errors with EOF or ErrUnexpectedEOF
			  OR got size-1 and wrong value

	testCodecFail( testing.T, codec, nonEmpty T )
	  t.Run("fail read") // does NOT use nonEmpty
		fails with r.Read() always returning an unknown error
		!!! NOT DOING THIS YET !!! Should *always* test no bytes read and EOF
		Not sure there's a reason to do this with read-short-buf testing
	  t.Run("fail write")
		w.Write fails wrinting nonEmpty with always failing Writer
		Replace this with writing to len(Append() - 1)
		exception when len(Append()) == 0
*/

/*
	New Structure

	testCodec/CodecVarying(t, codec, tests)
		bufCodec := ... normal/varying - changes which tests run
			OR maybe these methods differ?
		checkCodecMethods(t, bufCodec{...}, tests)

	checkCodecMethods(t, bufCodec, tests)
		// no t.Run("test_codec", ...)
		//   rely on test case names to distinguish normal/varying
		for tt := range tests
			t.Run(tt.name)
				// testCodec requires
				//   append/put/write(tt.value) == tt.data, size
				//   get/read(tt.data) == tt.value, exaust tt.data
				// testVarying (aka map) requires
				//   all pairs (vs having tt.data)
				//   buf := append nil/append existing/put/write(tt.value)
				//   get/read(buf) == tt.value
				//
				// Below documents what is common vs. distinct
				// Maybe just different bufferCodec implementations?
				// basic := just one "written value", test vs. Get/Read
				// varying: three written values, test vs. Get/Read
				//
				// !!! TEST THAT BUF IS NOT MODIFIED !!!
				// Could create anew whenever used, but then it's hard to test
				// that it's not modified.
				t.Run("append nil")
					basic:
						assert Append(nil, tt.value) == tt.data
					varying:
						buf := Append(nil, value)
						get:
							t.Run("=> get")
								got := bufCodec.Get(copy buf)
									got, gotSize := Get(buf)
									assert gotSize == len(buf)
									return got
								assert got == tt.value, and assert type
						read:
							t.Run("=> read")
								got := bufCodec.Read(copy buf)
									r := bytes.Reader(buf)
									got := Read(r)
									assert no error
									assert r.Len() == 0
									return got
								assert got == t.value, and assert type
				t.Run("append existing")
					basic:
						assert Append(header, tt.value) == [header, tt.data]
					varying:
						size := len(Append(nil, tt.value))
						buf := [header]
						buf := Append(buf, tt.value)
						assert header == buf[:header]
						assert size == len(buf) - len(header)
						buf = buf[header:]
						REPEAT get: and read: are the same
				t.Run("put")
					basic:
						size = len(Append(nil, tt.value))
						buf := [:size]
						assert Put(buf, tt.value) == size
						assert buf == tt.data
					varying:
						size := len(Append(nil, tt.value))
						buf := [:size]
						assert size == Put(buf, tt.value)
						REPEAT get: and read: are the same
				t.Run("put short buf")
					basic:
						size = len(Append(nil, tt.value))
						skip if size == 0
						buf := [:size+10000]
						Put(buf[:size-1], tt.value)
							assert panics
					varying:
						size = len(Append(nil, tt.value))
						skip if size == 0
						buf := [:size+10000]
						Put(buf[:size-1], tt.value)
							assert panics
				t.Run("get")
					basic:
						got, gotSize := Get(tt.data)
						assert gotSize == len(tt.data)
						assert got == tt.value, and assert type
					varying: NOT run separately, part of get/read: subtests
						got, gotSize := Get(buf)
						assert gotSize == len(buf)
						return got
						AFTER RETURN
						assert got == tt.value, and assert type
				!!! NOT DOING THIS YET !!! - separate test
					Should *always* test no bytes get and "EOF"
				t.Run("get short buf")
					basic:
						size := len(tt.data)
						skip if size == 0
						got, gotSize := Get(tt.data[:size-1])
							assert panics
							OR assert gotSize == size-1
								AND assert got != tt.value
					varying:
						buf := Append(nil, tt.value)
						size := len(buf)
						skip if size == 0
						got, gotSize := Get(buf[:size-1])
							assert panics
							OR assert gotSize == size-1
								AND assert got != tt.value
								AND assert types ==
				t.Run("write")
					basic:
						buf := bytes.Buffer
						Write(buf, tt.value)
						assert no error
						assert buf == tt.data
					varying:
						size := len(Append(nil, tt.value))
						buf := bytes.Buffer
						Write(buf, tt.value)
						assert no error
						assert size == buf.Len()
						REPEAT get: and read: are the same
				t.Run("write short buf")
					basic:
						size := len(Append(nil, tt.value))
						// internal buf is larger, but len=size-1
						writer := limitWriter(size-1)
						maybe iotest.TruncateWriter?
							OR two tests, writer silent and other errs
						Write(write, tt.Value)
						assert error (what kind?)
					varying: TODO
				t.Run("read")
					basic:
						r := bytes.Reader(tt.data)
						got := Read(r)
						assert no error
						assert r.Len() == 0
						assert got == tt.value, and assert type
					varying: NOT run separately, part of get/read: subtests
						r := bytes.Reader(buf)
						got := Read(r)
						assert no error
						assert r.Len() == 0
						AFTER RETURN
						assert got == tt.value, and assert type
				!!! NOT DOING THIS YET !!! - separate test
					Should *always* test no bytes read and EOF
				t.Run("read short buf")
					basic:
						size := len(tt.data)
						skip if size == 0
						r := bytes.Reader(tt.data[:size-1])
						got := Read(r)
						if error
							assert EOF (size == 1) or unexpected EOF
						else
							assert r.Len() == 0
							assert got != tt.value, and assert types ==
					varying:
						buf := Append(nil, tt.value)
						size := len(buf)
						skip if size == 0
						r := bytes.Reader(buf[:size-1])
						got, err := Read(r)
						if err
							assert EOF (size == 1) or unexpected EOF
						else
							assert r.Len() == 0
							assert got != tt.value, and assert types ==
*/

// Tests:
// - Append/Put/Write(testCase.value) == testCase.data
// - Get/Read(testCase.data) == testCase.value
// - len(Get/Put byte count return value) == len(testCase.data)
// - Get/Put panic when given a buffer that is 1 byte too short, or return incorrect values.
// - Read errs when given a stream that is 1 byte too short, or returns incorrect values.
//
//nolint:thelper
func testFooCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("test codec", func(t *testing.T) {
		testCodecGet(t, codec, tests)
		testCodecRead(t, codec, tests)
	})
}

//nolint:thelper
func testCodecGet[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("get", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				got, gotSize := codec.Get(tt.data)
				assert.Equal(t, len(tt.data), gotSize)
				assert.IsType(t, tt.value, got)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("get short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(tt.data)
				if size == 0 {
					return
				}
				// Should either panic, or read one fewer byte and get the wrong value back.
				//nolint:errcheck
				defer func() { recover() }()
				got, gotSize := codec.Get(tt.data[:size-1])
				assert.Equal(t, size-1, gotSize, "read too much data")
				assert.NotEqual(t, tt.value, got, "read value without full data")
			})
		}
	})
}

//nolint:thelper
func testCodecRead[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	t.Run("read", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				r := bytes.NewReader(tt.data)
				got, err := codec.Read(r)
				require.NoError(t, err)
				assert.Equal(t, 0, r.Len())
				assert.IsType(t, tt.value, got)
				assert.Equal(t, tt.value, got)
			})
		}
	})
	t.Run("read short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				size := len(tt.data)
				if size == 0 {
					return
				}
				// Should either err, or get the wrong value back.
				r := bytes.NewReader(tt.data[:size-1])
				got, err := codec.Read(r)
				if err != nil {
					assert.ErrorIs(t, err, errorForEOF(size-1))
					return
				}
				assert.Equal(t, 0, r.Len())
				assert.IsType(t, tt.value, got)
				assert.NotEqual(t, tt.value, got)
			})
		}
	})
}

// This tests Codecs whose encodings vary for the same input.
// Maps are the only current use case, because of their random iteration order.
// This function does not use testCase.data for this reason.
//
// Tests:
// - input == output, where input => Append/Put/Write => Get/Read => output
// - len(Append/Put/Write) are all equal
// - Get/Put panic when given a buffer that is 1 byte too short, or return incorrect values.
// - Read errs when given a stream that is 1 byte too short, or returns incorrect values.
//
//nolint:thelper
func testFooVaryingCodec[T any](t *testing.T, codec lexy.Codec[T], tests []testCase[T]) {
	bufCodec := bufferCodec[T]{codec}
	t.Run("short buf", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name+"-get", func(t *testing.T) {
				t.Parallel()
				buf := codec.Append(nil, tt.value)
				size := len(buf)
				if size == 0 {
					return
				}
				got, panicked := bufCodec.GetShortBuf(t, buf)
				if !panicked {
					assert.IsType(t, tt.value, got)
					assert.NotEqual(t, tt.value, got)
				}
			})
			t.Run(tt.name+"-read", func(t *testing.T) {
				t.Parallel()
				buf := codec.Append(nil, tt.value)
				size := len(buf)
				if size == 0 {
					return
				}
				got, err := bufCodec.ReadShortBuf(t, buf)
				if err != nil {
					assert.ErrorIs(t, err, errorForEOF(size-1))
					return
				}
				assert.IsType(t, tt.value, got)
				assert.NotEqual(t, tt.value, got)
			})
		}
	})
}

//nolint:thelper
func testRoundTripBuf[T any](t *testing.T, bufCodec bufferCodec[T], value T, name string, buf []byte) {
	t.Run(name+"-get", func(t *testing.T) {
		t.Parallel()
		got := bufCodec.Get(t, append([]byte{}, buf...))
		assert.IsType(t, value, got)
		assert.Equal(t, value, got)
	})
	t.Run(name+"-read", func(t *testing.T) {
		t.Parallel()
		got := bufCodec.Read(t, append([]byte{}, buf...))
		assert.IsType(t, value, got)
		assert.Equal(t, value, got)
	})
}

// A testing wrapper for Codec that deals only in []buf,
// and which tests some assertions while processing.
type bufferCodec[T any] struct {
	codec lexy.Codec[T]
}

//nolint:thelper
func (c bufferCodec[T]) Get(t *testing.T, buf []byte) T {
	got, gotSize := c.codec.Get(buf)
	assert.Equal(t, len(buf), gotSize)
	return got
}

//nolint:thelper,nonamedreturns
func (c bufferCodec[T]) GetShortBuf(t *testing.T, buf []byte) (_ T, panicked bool) {
	// Should either panic, or read one fewer byte and get the wrong value back.
	defer func() {
		//nolint:errcheck
		recover()
		panicked = true
	}()
	size := len(buf)
	got, gotSize := c.codec.Get(buf[:size-1])
	assert.Equal(t, size-1, gotSize, "read too much data")
	return got, false
}

//nolint:thelper
func (c bufferCodec[T]) Read(t *testing.T, buf []byte) T {
	r := bytes.NewReader(buf)
	got, err := c.codec.Read(r)
	require.NoError(t, err)
	assert.Equal(t, 0, r.Len())
	return got
}

//nolint:thelper
func (c bufferCodec[T]) ReadShortBuf(t *testing.T, buf []byte) (T, error) {
	size := len(buf)
	r := bytes.NewReader(buf[:size-1])
	got, err := c.codec.Read(r)
	if err != nil {
		var zero T
		return zero, err
	}
	assert.Equal(t, 0, r.Len())
	return got, nil
}

var (
	errRead  = errors.New("failed to read")
	errWrite = errors.New("failed to write")
)

// Tests:
// - Codec.Read() fails when reading from a failing io.Reader.
// - Codec.Write() fails when writing nonEmpty to a failing io.Writer.
func testCodecFail[T any](t *testing.T, codec lexy.Codec[T], nonEmpty T) {
	t.Helper()
	t.Run("fail read", func(t *testing.T) {
		t.Parallel()
		_, err := codec.Read(iotest.ErrReader(errRead))
		assert.Error(t, err)
	})
	t.Run("fail write", func(t *testing.T) {
		t.Parallel()
		err := codec.Write(failWriter{}, nonEmpty)
		assert.Error(t, err)
	})
}

type failWriter struct{}

var _ io.Writer = failWriter{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errWrite
}
