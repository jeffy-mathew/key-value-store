package repository

import (
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyValueStore(t *testing.T) {
	var (
		key   = "key"
		value = []byte("value")
	)
	logger := zerolog.New(os.Stdout)

	t.Run("NewKeyValueStore", func(t *testing.T) {
		store, err := NewKeyValueStore(logger)
		require.NoError(t, err)

		if store.data == nil {
			t.Error("Store data map not initialized")
		}
	})

	t.Run("Set", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger)

		err := store.Set(context.Background(), key, value)
		require.NoError(t, err)

		got, ok, err := store.Get(context.Background(), key)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, value, got)
	})

	t.Run("Get", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger)
		store.data = map[string][]byte{
			key: value,
		}

		err := store.Set(context.Background(), "key", []byte("new-value"))
		require.NoError(t, err)
	})

	t.Run("Set and Get", func(t *testing.T) {
		store, _ := NewKeyValueStore(logger)

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
		store, _ := NewKeyValueStore(logger)

		ctx := context.Background()

		assert.NoError(t, store.Set(ctx, key, value))
		err := store.Delete(ctx, key)
		assert.NoError(t, err)

		_, exists, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
