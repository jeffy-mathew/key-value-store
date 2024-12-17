package repository

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/gob"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

//go:embed testdata/benchmark_data.gob
var BenchmarkData []byte

type BenchmarkSuite struct {
	store    Store
	testKeys []string
}

func generateValue(b *testing.B, size int) []byte {
	b.Helper()
	return []byte(strings.Repeat("v", size))
}

func getSeedData(b *testing.B) map[string][]byte {
	b.Helper()

	// Create a decoder for the embedded data
	decoder := gob.NewDecoder(bytes.NewReader(BenchmarkData))

	// Create a struct to hold the decoded data
	var data struct {
		Store map[string][]byte
	}

	// Decode the data
	if err := decoder.Decode(&data); err != nil {
		b.Fatalf("Failed to decode benchmark data: %v", err)
	}

	return data.Store
}

func setupBenchmark(b *testing.B) *BenchmarkSuite {
	b.Helper()

	logger := zerolog.New(zerolog.NewConsoleWriter())

	store, err := NewKeyValueStore(logger)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}

	// Seed the store with benchmark data
	seedData := getSeedData(b)
	store.Seed(seedData)

	// Extract test keys for benchmarking
	keys := make([]string, 0, len(seedData))
	for k := range seedData {
		keys = append(keys, k)
	}

	return &BenchmarkSuite{
		store:    store,
		testKeys: keys,
	}
}

// BenchmarkDirectWrites tests write operations directly on the store
func BenchmarkDirectWrites(b *testing.B) {
	suite := setupBenchmark(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("key-%d", time.Now().UnixNano())
			value := generateValue(b, 16+rand.Intn(985))
			if err := suite.store.Set(context.Background(), key, value); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDirectReads tests read operations directly on the store
func BenchmarkDirectReads(b *testing.B) {
	suite := setupBenchmark(b)

	numKeys := len(suite.testKeys)
	if numKeys == 0 {
		b.Fatal("No test keys available")
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// each goroutine should maintain its own counter for key rotation
		keyIndex := 0
		for pb.Next() {
			key := suite.testKeys[keyIndex]
			_, exists, err := suite.store.Get(context.Background(), key)
			if err != nil {
				b.Fatal(err)
			}
			if !exists {
				b.Fatalf("Key not found: %s", key)
			}
			// rotate through keys
			keyIndex = (keyIndex + 1) % numKeys
		}
	})
}

// BenchmarkMixedDirectOperations tests a combination of read and write operations
func BenchmarkMixedDirectOperations(b *testing.B) {
	suite := setupBenchmark(b)

	if err := suite.store.Set(context.Background(), "key", []byte("value")); err != nil {
		b.Fatalf("Failed to set initial data: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// each goroutine gets its own counter
		localCounter := 0
		for pb.Next() {
			if localCounter%3 == 0 { // 33% writes, 67% reads
				writeKey := fmt.Sprintf("key-%d", time.Now().UnixNano())
				writeValue := []byte(fmt.Sprintf("value-%d", time.Now().UnixNano()))
				if err := suite.store.Set(context.Background(), writeKey, writeValue); err != nil {
					b.Fatal(err)
				}
			} else {
				key := suite.testKeys[localCounter%len(suite.testKeys)]
				_, exists, err := suite.store.Get(context.Background(), key)
				if err != nil {
					b.Fatal(err)
				}
				if !exists {
					b.Fatal("Key not found")
				}
			}
			localCounter++
		}
	})
}

// BenchmarkHighConcurrencyDirectOperations tests the system under very high concurrent load
func BenchmarkHighConcurrencyDirectOperations(b *testing.B) {
	suite := setupBenchmark(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("key-%d", time.Now().UnixNano())
			value := []byte(fmt.Sprintf("value-%d", time.Now().UnixNano()))
			if err := suite.store.Set(context.Background(), key, value); err != nil {
				b.Fatal(err)
			}
		}
	})
}
