package lexy_test

import (
	"bytes"
	"fmt"
	"sort"

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

func (db *DB) insert(i int, entry Entry) {
	db.entries = append(db.entries, Entry{})
	copy(db.entries[i+1:], db.entries[i:])
	db.entries[i] = entry
}

func (db *DB) search(entry Entry) (int, bool) {
	index := sort.Search(len(db.entries), func(i int) bool {
		return bytes.Compare(entry.Key, db.entries[i].Key) <= 0
	})
	if index < len(db.entries) && bytes.Equal(entry.Key, db.entries[index].Key) {
		return index, true
	}
	return index, false
}

func (db *DB) Put(key []byte, value int) error {
	entry := Entry{key, value}
	if i, found := db.search(entry); found {
		db.entries[i] = entry
	} else {
		db.insert(i, entry)
	}
	return nil
}

// Returns Entries, in order, such that (begin <= entry.Key < end).
func (db *DB) Range(begin, end []byte) ([]Entry, error) {
	a, _ := db.search(Entry{begin, 0})
	b, _ := db.search(Entry{end, 0})
	return db.entries[a:b], nil
}

// END TOY DB IMPLEMENTATION

// BEGIN KEY CODEC

var (
	wordsCodec = lexy.TerminateIfNeeded(lexy.SliceOf(lexy.String()))
	costCodec  = lexy.Int32()
	keyCodec   = KeyCodec{}
)

type KeyCodec struct{}

func (KeyCodec) Append(buf []byte, key UserKey) []byte {
	buf = costCodec.Append(buf, key.cost)
	return wordsCodec.Append(buf, key.words)
}

func (KeyCodec) Put(buf []byte, key UserKey) []byte {
	buf = costCodec.Put(buf, key.cost)
	return wordsCodec.Put(buf, key.words)
}

func (KeyCodec) Get(buf []byte) (UserKey, []byte) {
	cost, buf := costCodec.Get(buf)
	words, buf := wordsCodec.Get(buf)
	return UserKey{words, cost}, buf
}

func (KeyCodec) RequiresTerminator() bool {
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
	return db.realDB.Put(keyCodec.Append(nil, key), value)
}

// Returns Entries, in order, such that (begin <= entry.Key < end).
func (db *UserDB) Range(begin, end UserKey) ([]UserEntry, error) {
	beginBytes := keyCodec.Append(nil, begin)
	endBytes := keyCodec.Append(nil, end)
	dbEntries, err := db.realDB.Range(beginBytes, endBytes)
	if err != nil {
		return nil, err
	}
	userEntries := make([]UserEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		userKey, _ := keyCodec.Get(dbEntry.Key)
		userEntries[i] = UserEntry{userKey, dbEntry.Value}
	}
	return userEntries, nil
}

// END USER DB ABSTRACTION

// ExampleRangeQuery shows how a range query might be implemented.
func Example_rangeQuery() {
	var userDB UserDB
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
