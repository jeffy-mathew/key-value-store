//go:build integration

package tests

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"codesignal/internal/config"
	"codesignal/internal/repository"
	"codesignal/internal/router"
	"codesignal/internal/store"
)

type BenchData struct {
	Store map[string][]byte
}

type BenchmarkSuite struct {
	server   *httptest.Server
	store    repository.Store
	testKeys []string
}

func generateValue(b *testing.B, size int) string {
	b.Helper()
	return strings.Repeat("v", size)
}

func getSeedData(b *testing.B) map[string][]byte {
	b.Helper()
	file, err := os.Open("testdata/benchmark_data.gob")
	if err != nil {
		b.Fatalf("Failed to open benchmark data file: %v", err)
	}

	b.Cleanup(func() {
		file.Close()
	})

	var data BenchData
	err = gob.NewDecoder(file).Decode(&data)
	require.NoError(b, err)

	return data.Store
}

func setupTestServer(b *testing.B) *BenchmarkSuite {
	b.Helper()
	log := zerolog.New(io.Discard)
	store, err := repository.NewKeyValueStore(log)
	require.NoError(b, err)

	seedData := getSeedData(b)
	store.Seed(seedData)

	keys := make([]string, 0, len(seedData))
	for k := range seedData {
		keys = append(keys, k)
	}

	cfg := &config.Config{
		MaxKeyLength: 100,
		MaxValueSize: 1024, // 1MB
	}

	handler := router.New(log, store, cfg)
	server := httptest.NewServer(handler)
	return &BenchmarkSuite{
		server:   server,
		store:    store,
		testKeys: keys,
	}
}

func BenchmarkSetAPI(b *testing.B) {
	suite := setupTestServer(b)
	b.Cleanup(suite.server.Close)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Pre-generate random keys and values
	keys := make([]string, b.N)
	values := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", time.Now().UnixNano())
		values[i] = generateValue(b, 16+rand.Intn(985))
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		kv := store.KeyValue{
			Key:   keys[i],
			Value: values[i],
		}

		jsonData, err := json.Marshal(kv)
		require.NoError(b, err)

		req, err := http.NewRequest(http.MethodPost, suite.server.URL+"/key", bytes.NewBuffer(jsonData))
		require.NoError(b, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(b, err)
		_, err = io.Copy(io.Discard, resp.Body)
		require.NoError(b, err)
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
			b.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func BenchmarkGetAPI(b *testing.B) {
	suite := setupTestServer(b)
	b.Cleanup(suite.server.Close)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", b.N%1000)
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/key/%s", suite.server.URL, key), nil)
		require.NoError(b, err)

		resp, err := client.Do(req)
		require.NoError(b, err)
		_, err = io.Copy(io.Discard, resp.Body)
		require.NoError(b, err)
		require.Equal(b, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func BenchmarkDeleteAPI(b *testing.B) {
	suite := setupTestServer(b)
	b.Cleanup(suite.server.Close)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", b.N%1000)
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/key/%s", suite.server.URL, key), nil)
		require.NoError(b, err)

		resp, err := client.Do(req)
		require.NoError(b, err)
		_, err = io.Copy(io.Discard, resp.Body)
		require.NoError(b, err)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			b.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}
