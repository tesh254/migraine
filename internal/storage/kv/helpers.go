package kv

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
)

// Store represents a BadgerDB key-value store
type Store struct {
	db *badger.DB
}

// New creates a new Store instance
func New(db *badger.DB) *Store {
	return &Store{db: db}
}

// Set stores a value for a key
func (s *Store) Set(key string, value interface{}) error {
	return s.db.Update(func(txn *badger.Txn) error {
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), bytes)
	})
}

// Get retrieves a value for a key
func (s *Store) Get(key string, value interface{}) error {
	return s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, value)
		})
	})
}

// Delete removes a key
func (s *Store) Delete(key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// List returns all keys with a given prefix
func (s *Store) List(prefix string) ([]string, error) {
	var keys []string
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key())
			keys = append(keys, key)
		}
		return nil
	})
	return keys, err
}
