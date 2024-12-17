package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Data struct {
	Store map[string][]byte
}

func generateValue(size int) []byte {
	return []byte(strings.Repeat("v", size))
}

func main() {
	data := Data{
		Store: make(map[string][]byte),
	}

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		// Vary value size between 16 bytes and 1KB
		valueSize := 16 + (i % 985) // max size will be 1000 bytes
		data.Store[key] = generateValue(valueSize)
	}

	// Encode to gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		log.Fatalf("Failed to encode data: %v", err)
	}

	// Write to file
	outFile := filepath.Join(".assets", "benchmark_data.gob")
	if err := os.WriteFile(outFile, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write gob file: %v", err)
	}

	log.Printf("Successfully created %s with %d keys", outFile, len(data.Store))
}
