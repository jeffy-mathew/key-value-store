package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyValueStore(t *testing.T) {
	var (
		key   = "key"
		value = []byte("value")
	)

	tmpFile := "/var/tmp/test_store.json"
	t.Cleanup(func() {
		os.Remove(tmpFile)
	})

	Opts := Opts{
		SyncInterval: 100 * time.Millisecond,
		DataFile:     tmpFile,
	}
	logger := zerolog.New(os.Stdout)

	t.Run("NewKeyValueStore", func(t *testing.T) {
		store, err := NewKeyValueStore(logger, Opts)
		if err != nil {
			t.Fatalf("Failed to create new store: %v", err)
		}
		defer store.Close()

		if store.data == nil {
			t.Error("Store data map not initialized")
		}
	})

	t.Run("Set", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger, Opts)
		t.Cleanup(func() {
			_ = store.Close()
		})

		err := store.Set(context.Background(), key, value)
		require.NoError(t, err)

		got, ok, err := store.Get(context.Background(), key)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, value, got)
	})

	t.Run("Get", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger, Opts)
		t.Cleanup(func() {
			_ = store.Close()
		})

		store.data = map[string][]byte{
			key: value,
		}

		err := store.Set(context.Background(), "key", []byte("new-value"))
		require.NoError(t, err)
	})

	t.Run("Set and Get", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger, Opts)
		defer store.Close()

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")

		err := store.Set(ctx, key, value)
		if err != nil {
			t.Errorf("Failed to set value: %v", err)
		}

		retrieved, exists, err := store.Get(ctx, key)
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, value, retrieved)
	})

	t.Run("Delete", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger, Opts)
		defer store.Close()

		ctx := context.Background()

		assert.NoError(t, store.Set(ctx, key, value))
		err := store.Delete(ctx, key)
		assert.NoError(t, err)

		_, exists, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Persistence", func(t *testing.T) {
		store1, _ := NewKeyValueStore(logger, Opts)
		ctx := context.Background()

		store1.Set(ctx, key, value)
		require.NoError(t, store1.Close())

		store2, _ := NewKeyValueStore(logger, Opts)
		t.Cleanup(func() {
			store2.Close()
		})

		retrieved, exists, err := store2.Get(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, value, retrieved)
	})
}
