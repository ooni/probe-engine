// Package kvstore contains key-value stores
package kvstore

import "errors"

// MemoryKeyValueStore is an in-memory key-value store
type MemoryKeyValueStore map[string]string

// NewMemoryKeyValueStore creates a new in-memory key-value store
func NewMemoryKeyValueStore() MemoryKeyValueStore {
	return MemoryKeyValueStore{}
}

// Get returns a key from the key value store
func (kvs MemoryKeyValueStore) Get(key string) (string, error) {
	var (
		err   error
		ok    bool
		value string
	)
	value, ok = kvs[key]
	if !ok {
		err = errors.New("no such key")
	}
	return value, err
}

// Set sets a key into the key value store
func (kvs MemoryKeyValueStore) Set(key, value string) error {
	kvs[key] = value
	return nil
}
