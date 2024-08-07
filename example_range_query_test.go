package lexy_test

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/phiryll/lexy"
)

// BEGIN TOY DB IMPLEMENTATION

type DB struct {
	entries []Entry // sort order by Entry.key is maintained
}

type Entry struct {
	Key   []byte
	Value int // value type is unimportant for this example
}

func cmpEntries(a, b Entry) int { return bytes.Compare(a.Key, b.Key) }

func (db *DB) insert(i int, entry Entry) {
	entries := append(db.entries, Entry{})
	copy(entries[i+1:], entries[i:])
	entries[i] = entry
	db.entries = entries
}

func (db *DB) Put(key []byte, value int) error {
	entry := Entry{key, value}
	if i, found := slices.BinarySearchFunc(db.entries, entry, cmpEntries); found {
		db.entries[i] = entry
	} else {
		db.insert(i, entry)
	}
	return nil
}

// Returns Entries, in order, such that (begin <= entry.Key < end)
func (db *DB) Range(begin, end []byte) ([]Entry, error) {
	a, _ := slices.BinarySearchFunc(db.entries, Entry{begin, 0}, cmpEntries)
	b, _ := slices.BinarySearchFunc(db.entries, Entry{end, 0}, cmpEntries)
	return db.entries[a:b], nil
}

// END TOY DB IMPLEMENTATION

// BEGIN KEY CODEC

var (
	wordsCodec = lexy.Terminate(lexy.SliceOf[[]string](lexy.String[string]()))
	costCodec  = lexy.Int32[int32]()
	keyCodec   = KeyCodec{}
)

type KeyCodec struct{}

func (c KeyCodec) Read(r io.Reader) (UserKey, error) {
	var zero UserKey
	cost, err := costCodec.Read(r)
	if err != nil {
		return zero, err
	}
	words, err := wordsCodec.Read(r)
	if err != nil {
		return zero, lexy.UnexpectedIfEOF(err)
	}
	return UserKey{words, cost}, nil
}

func (c KeyCodec) Write(w io.Writer, key UserKey) error {
	if err := costCodec.Write(w, key.cost); err != nil {
		return err
	}
	return wordsCodec.Write(w, key.words)
}

func (c KeyCodec) RequiresTerminator() bool {
	return false
}

// END KEY CODEC

// BEGIN USER DB ABSTRACTION

type UserKey struct {
	words []string
	cost  int32
}

func (k UserKey) String() string {
	return fmt.Sprintf("{%d, %v}", k.cost, k.words)
}

type UserDB struct {
	realDB DB
}

type UserEntry struct {
	Key   UserKey
	Value int
}

func (db *UserDB) Put(key UserKey, value int) error {
	var buf bytes.Buffer
	if err := keyCodec.Write(&buf, key); err != nil {
		return err
	}
	return db.realDB.Put(buf.Bytes(), value)
}

// Returns Entries, in order, such that (begin <= entry.Key < end)
func (db *UserDB) Range(begin, end UserKey) ([]UserEntry, error) {
	var buf bytes.Buffer
	if err := keyCodec.Write(&buf, begin); err != nil {
		return nil, err
	}
	beginBytes := bytes.Clone(buf.Bytes())
	buf.Reset()
	if err := keyCodec.Write(&buf, end); err != nil {
		return nil, err
	}
	endBytes := bytes.Clone(buf.Bytes())
	dbEntries, err := db.realDB.Range(beginBytes, endBytes)
	if err != nil {
		return nil, err
	}
	userEntries := make([]UserEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		userKey, err := keyCodec.Read(bytes.NewReader(dbEntry.Key))
		if err != nil {
			return nil, err
		}
		userEntries[i] = UserEntry{userKey, dbEntry.Value}
	}
	return userEntries, nil
}

// END USER DB ABSTRACTION

// Example (RangeQuery) shows how a range query might be implemented.
func Example_rangeQuery() {
	userDB := UserDB{}
	for _, item := range []struct {
		cost  int32
		words []string
		value int
	}{
		// In sort order for clarity: key.Cost, then key.Words
		{1, []string{"not"}, 0},
		{1, []string{"not", "the"}, 0},
		{1, []string{"not", "the", "end"}, 0},
		{1, []string{"now"}, 0},

		{2, []string{"iffy", "proposal"}, 0},
		{2, []string{"in"}, 0},
		{2, []string{"in", "cahoots"}, 0},
		{2, []string{"in", "sort"}, 0},
		{2, []string{"in", "sort", "order"}, 0},
		{2, []string{"integer", "sort"}, 0},
	} {
		err := userDB.Put(UserKey{item.words, item.cost}, item.value)
		if err != nil {
			panic(err)
		}
	}

	printRange := func(low, high UserKey) {
		fmt.Printf("Range: %s -> %s\n", low.String(), high.String())
		entries, err := userDB.Range(low, high)
		if err != nil {
			panic(err)
		}
		for _, userEntry := range entries {
			fmt.Println(userEntry.Key.String())
		}
	}

	printRange(UserKey{[]string{"an"}, -1000},
		UserKey{[]string{"empty", "result"}, 1})
	printRange(UserKey{[]string{}, 1},
		UserKey{[]string{"not", "the", "beginning"}, 1})
	printRange(UserKey{[]string{"nouns", "are", "words"}, 1},
		UserKey{[]string{"in", "sort", "disorder"}, 2})
	// Output:
	// Range: {-1000, [an]} -> {1, [empty result]}
	// Range: {1, []} -> {1, [not the beginning]}
	// {1, [not]}
	// {1, [not the]}
	// Range: {1, [nouns are words]} -> {2, [in sort disorder]}
	// {1, [now]}
	// {2, [iffy proposal]}
	// {2, [in]}
	// {2, [in cahoots]}
	// {2, [in sort]}
}
