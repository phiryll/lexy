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

func (db *DB) Put(key []byte, value int) {
	entry := Entry{key, value}
	if i, found := slices.BinarySearchFunc(db.entries, entry, cmpEntries); found {
		db.entries[i] = entry
	} else {
		db.entries = slices.Insert(db.entries, i, entry)
	}
}

// Returns Entries, in order, such that (begin <= entry.Key < end)
func (db *DB) Range(begin, end []byte) []Entry {
	a, _ := slices.BinarySearchFunc(db.entries, Entry{begin, 0}, cmpEntries)
	b, _ := slices.BinarySearchFunc(db.entries, Entry{end, 0}, cmpEntries)
	return db.entries[a:b]
}

// END TOY DB IMPLEMENTATION

// BEGIN KEY CODEC

var (
	wordsCodec = lexy.Terminate(lexy.SliceOf[[]string](lexy.String[string]()))
	costCodec  = lexy.Int[int32]()
	keyCodec   = KeyCodec{}
)

type KeyCodec struct{}

func (c KeyCodec) Read(r io.Reader) (UserKey, error) {
	cost, _ := costCodec.Read(r)
	words, _ := wordsCodec.Read(r)
	return UserKey{words, cost}, nil
}

func (c KeyCodec) Write(w io.Writer, key UserKey) error {
	costCodec.Write(w, key.cost)
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

func (db *UserDB) Put(key UserKey, value int) {
	var buf bytes.Buffer
	keyCodec.Write(&buf, key)
	db.realDB.Put(buf.Bytes(), value)
}

// Returns Entries, in order, such that (begin <= entry.Key < end)
func (db *UserDB) Range(begin, end UserKey) []UserEntry {
	var buf bytes.Buffer
	keyCodec.Write(&buf, begin)
	beginBytes := bytes.Clone(buf.Bytes())
	buf.Reset()
	keyCodec.Write(&buf, end)
	endBytes := bytes.Clone(buf.Bytes())
	dbEntries := db.realDB.Range(beginBytes, endBytes)
	userEntries := make([]UserEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		userKey, _ := keyCodec.Read(bytes.NewReader(dbEntry.Key))
		userEntries[i] = UserEntry{userKey, dbEntry.Value}
	}
	return userEntries
}

// END USER DB ABSTRACTION

// Example (RangeQuery) shows how a range query might be implemented.
// Because this example is so long, error handling has been removed.
// DON'T DO THIS!
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
		userDB.Put(UserKey{item.words, item.cost}, item.value)
	}

	printRange := func(low, high UserKey) {
		fmt.Printf("Range: %s -> %s\n", low.String(), high.String())
		for _, userEntry := range userDB.Range(low, high) {
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
