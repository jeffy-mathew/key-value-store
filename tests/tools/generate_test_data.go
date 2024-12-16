package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Data struct {
	Store map[string][]byte
}

func generateValue(size int) []byte {
	return []byte(strings.Repeat("v", size))
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to get current file path")
	}
	dir := filepath.Dir(filename)

	data := Data{
		Store: make(map[string][]byte),
	}

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		// Vary value size between 16 bytes and 1KB
		valueSize := 16 + (i % 985) // max size will be 1000 bytes
		data.Store[key] = generateValue(valueSize)
	}

	// Add some special keys
	data.Store["empty"] = []byte("")
	data.Store["small"] = []byte("small value")
	data.Store["medium"] = generateValue(500)
	data.Store["large"] = generateValue(1000)
	data.Store["session:active"] = []byte("session data")
	data.Store["user:1234"] = []byte("user profile")
	data.Store["config:app"] = []byte("app configuration")
	data.Store["cache:temp"] = []byte("cached data")

	// Encode to gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		log.Fatalf("Failed to encode data: %v", err)
	}

	// Write to file
	outFile := filepath.Join(dir, "../testdata/benchmark_data.gob")
	if err := os.WriteFile(outFile, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write gob file: %v", err)
	}

	log.Printf("Successfully created %s with %d keys", outFile, len(data.Store))
}
