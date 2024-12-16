package repository

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Store represents the interface for key-value store operations.
type Store interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
	Close() error
}

// Config holds the configuration for the store
type Config struct {
	SyncInterval time.Duration
	DataFile     string
}

// KeyValueStore implements the Store interface with persistence.
type KeyValueStore struct {
	data   map[string][]byte
	mu     sync.RWMutex
	log    zerolog.Logger
	config Config
	done   chan struct{}
}

// Data represents the structure for persistence.
type Data struct {
	Store map[string][]byte
}

// NewKeyValueStore creates a new instance of KeyValueStore
func NewKeyValueStore(log zerolog.Logger, config Config) (*KeyValueStore, error) {
	kvs := &KeyValueStore{
		data:   make(map[string][]byte),
		log:    log,
		config: config,
		done:   make(chan struct{}),
	}

	if err := kvs.loadFromDisk(); err != nil {
		return nil, err
	}

	go kvs.startSync()

	return kvs, nil
}

func (k *KeyValueStore) Set(ctx context.Context, key string, value []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = value
	return nil
}

func (k *KeyValueStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	value, exists := k.data[key]
	return value, exists, nil
}

func (k *KeyValueStore) Delete(ctx context.Context, key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.data, key)
	return nil
}

func (k *KeyValueStore) Close() error {
	close(k.done)
	return k.syncToDisk()
}

// startSync starts a goroutine to sync data to dis at SyncInterval.
func (k *KeyValueStore) startSync() {
	ticker := time.NewTicker(k.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := k.syncToDisk(); err != nil {
				k.log.Error().Err(err).Msg("failed to sync data to disk")
			}
		case <-k.done:
			return
		}
	}
}

func (k *KeyValueStore) syncToDisk() error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	file, err := os.Create(k.config.DataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	data := Data{
		Store: k.data,
	}

	return gob.NewEncoder(file).Encode(data)
}

func (k *KeyValueStore) loadFromDisk() error {
	data, err := os.ReadFile(k.config.DataFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var stored Data
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&stored); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	k.data = stored.Store
	return nil
}
