//go:generate mockgen -source=repository.go -destination=./mock/repository_mock.go -package=mock
package repository

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

// Store represents the interface for key-value store operations.
type Store interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

// KeyValueStore implements the Store interface with persistence.
type KeyValueStore struct {
	data map[string][]byte
	mu   *sync.RWMutex
	log  zerolog.Logger
}

// Data represents the structure for persistence.
type Data struct {
	Store map[string][]byte
}

// NewKeyValueStore creates a new instance of KeyValueStore
func NewKeyValueStore(log zerolog.Logger) (*KeyValueStore, error) {
	kvs := &KeyValueStore{
		mu:   &sync.RWMutex{},
		data: make(map[string][]byte),
		log:  log,
	}
	return kvs, nil
}

// Seed populates the store with initial data, used in tests.
// TODO: move this to a persistence layer
func (k *KeyValueStore) Seed(data map[string][]byte) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data = data
}

// Set sets a key-value pair in the store.
func (k *KeyValueStore) Set(ctx context.Context, key string, value []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = value
	return nil
}

// Get retrieves a value from the store by key.
func (k *KeyValueStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	value, exists := k.data[key]
	return value, exists, nil
}

// Delete deletes a key from the store.
func (k *KeyValueStore) Delete(ctx context.Context, key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.data, key)
	return nil
}
