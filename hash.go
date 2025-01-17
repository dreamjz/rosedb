package rosedb

import (
	"bytes"
	"github.com/flower-corp/rosedb/logfile"
)

const hashHeaderSize = 10

// HSet sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
// If field already exists in the hash, it is overwritten.
// Return num of elements in hash of the specified key.
func (db *RoseDB) HSet(key, field, value []byte) error {
	// If the existed value is the same as the set value, nothing will be done.
	oldVal := db.HGet(key, field)
	if bytes.Compare(oldVal, value) == 0 {
		return nil
	}

	db.hashIndex.mu.Lock()
	defer db.hashIndex.mu.Unlock()

	hashKey := db.encodeKey(key, field)
	entry := &logfile.LogEntry{Key: hashKey, Value: value}
	if _, err := db.writeLogEntry(entry, Hash); err != nil {
		return err
	}

	db.hashIndex.indexes.HSet(string(key), string(field), value)
	return nil
}

// HGet returns the value associated with field in the hash stored at key.
func (db *RoseDB) HGet(key, field []byte) []byte {
	db.hashIndex.mu.RLock()
	defer db.hashIndex.mu.RUnlock()
	return db.hashIndex.indexes.HGet(string(key), string(field))
}

// HDel removes the specified fields from the hash stored at key.
// Specified fields that do not exist within this hash are ignored.
// If key does not exist, it is treated as an empty hash and this command returns false.
func (db *RoseDB) HDel(key []byte, fields ...[]byte) (int, error) {
	db.hashIndex.mu.Lock()
	defer db.hashIndex.mu.Unlock()

	var count int
	for _, field := range fields {
		hashKey := db.encodeKey(key, field)
		entry := &logfile.LogEntry{Key: hashKey, Type: logfile.TypeDelete}
		if _, err := db.writeLogEntry(entry, Hash); err != nil {
			return 0, err
		}
		ok := db.hashIndex.indexes.HDel(string(key), string(field))
		if ok {
			count++
		}
	}
	return count, nil
}
