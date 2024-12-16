// Package store provides a service for managing a key-value store.
//
// This package includes the definition of a Service struct which
// encapsulates methods for setting, getting, and deleting keys in
// the key-value store.
package store

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"

	"codesignal/internal/repository"
)

// KeyValue represents a key-value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Response represents the API response
type Response struct {
	Message string `json:"message"`
}

// Service for managing a key value store.
type Service struct {
	log   zerolog.Logger
	store repository.Store
}

// NewService returns a new instance of Service.
func NewService(log zerolog.Logger, store repository.Store) *Service {
	return &Service{
		log:   log,
		store: store,
	}
}

func (s *Service) SetKey(w http.ResponseWriter, req *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(r.Context(),&kv); err != nil {
		s.log.Error().Err(err).Msg("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid request body"})
		return
	}

	// Check if key exists
	if _, exists := s.store.Get(r.Context(),, kv.Key); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Message: "key already exists"})
		return
	}

	if err := s.store.Set(r.Context(), kv.Key, kv.Value); err != nil {
		s.log.Error().Err(err).Msg("failed to set key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to set key"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "key created successfully"})
}

func (s *Service) GetKey(w http.ResponseWriter, req *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(r.Context(),&kv); err != nil {
		s.log.Error().Err(err).Msg("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid request body"})
		return
	}

	value, exists := s.store.Get(key)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Message: "key not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(KeyValue{Key: key, Value: value})
}

func (s *Service) DeleteKey(w http.ResponseWriter, req *http.Request) {
	key := extractKey(r.URL.Path)
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid key"})
		return
	}

	if _, exists := s.store.Get(r.Context(), key); !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Message: "key not found"})
		return
	}

	if err := s.store.Delete(r.Context(), key); err != nil {
		s.log.Error().Err(err).Msg("failed to delete key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to delete key"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "key deleted successfully"})
}
