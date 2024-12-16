package tests

import (
	"context"
	_ "embed"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"codesignal/internal/repository"

	"github.com/rs/zerolog"
)

//go:embed testdata/benchmark_data.gob
var BenchmarkData []byte

type BenchmarkSuite struct {
	srv      *httptest.Server
	client   *http.Client
	store    repository.Store
	dataFile string
	testKeys []string
}

func generateValue(b *testing.B, size int) []byte {
	b.Helper()
	return []byte(strings.Repeat("v", size))
}

func copyGobData(b *testing.B) string {
	b.Helper()

	//temp test file to run benchmark
	dataFile := fmt.Sprintf("testdata/benchmark_data_%d.gob", time.Now().UnixNano())
	if err := os.WriteFile(dataFile, BenchmarkData, 0644); err != nil {
		b.Fatalf("Failed to write benchmark data: %v", err)
	}
	return dataFile
}

func setupBenchmark(b *testing.B) *BenchmarkSuite {
	b.Helper()

	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Copy the embedded gob data to a temporary file
	dataFile := copyGobData(b)

	opts := repository.Opts{
		SyncInterval: time.Second * 5,
		DataFile:     dataFile,
	}

	store, err := repository.NewKeyValueStore(logger, opts)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}

	// Generate sequential keys for benchmarking
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	r := http.NewServeMux()
	srv := httptest.NewServer(r)

	return &BenchmarkSuite{
		srv:      srv,
		client:   &http.Client{},
		store:    store,
		dataFile: dataFile,
		testKeys: keys,
	}
}

func (s *BenchmarkSuite) teardown() {
	s.srv.Close()
	s.store.Close()
	// clean up the temporary gob file
	if err := os.Remove(s.dataFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove data file %s: %v\n", s.dataFile, err)
	}
}

// BenchmarkDirectWrites tests write operations directly on the store
func BenchmarkDirectWrites(b *testing.B) {
	suite := setupBenchmark(b)
	b.Cleanup(suite.teardown)

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
	b.Cleanup(suite.teardown)

	numKeys := len(suite.testKeys)
	if numKeys == 0 {
		b.Fatal("No test keys available")
	}

	b.ResetTimer()
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
	b.Cleanup(suite.teardown)

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
	b.Cleanup(suite.teardown)

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
