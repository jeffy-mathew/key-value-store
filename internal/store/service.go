// Package store provides a service for managing a key-value store.
//
// This package includes the definition of a Service struct which
// encapsulates methods for setting, getting, and deleting keys in
// the key-value store.
package store

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"

	"codesignal/internal/repository"
)

// KeyValue represents a key-value pair.
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StatusCode represents custom application status code for the API response.
type StatusCode int

// Status codes for the key-value store operations
const (
	StatusSuccess      StatusCode = 1000
	StatusKeyNotFound  StatusCode = 1001
	StatusKeyExists    StatusCode = 1002
	StatusInvalidKey   StatusCode = 1003
	StatusInvalidValue StatusCode = 1004
	StatusStorageError StatusCode = 1005
	StatusInvalidJSON  StatusCode = 1006
)

// Response represents the API response
type Response struct {
	Message    string     `json:"message"`
	StatusCode StatusCode `json:"status_code"`
	Data       KeyValue   `json:"data,omitempty"`
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

func (s *Service) SetKey(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		s.log.Error().Err(err).Msg("failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid request body", StatusCode: StatusInvalidJSON})
		return
	}

	_, exists, err := s.store.Get(r.Context(), kv.Key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Message: "key already exists", StatusCode: StatusKeyExists})
		return
	}

	if err := s.store.Set(r.Context(), kv.Key, []byte(kv.Value)); err != nil {
		s.log.Error().Err(err).Msg("failed to set key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to set key", StatusCode: StatusStorageError})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "key created successfully", StatusCode: StatusSuccess})
}

func (s *Service) GetKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	key := params.ByName("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid key", StatusCode: StatusInvalidKey})
		return
	}

	kv, exists, err := s.store.Get(r.Context(), key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Message: "key not found", StatusCode: StatusKeyNotFound})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "key found successfully", StatusCode: StatusSuccess,
		Data: KeyValue{
			Key:   key,
			Value: string(kv),
		}})
}

func (s *Service) DeleteKey(w http.ResponseWriter, req *http.Request) {
	params := httprouter.ParamsFromContext(req.Context())

	key := params.ByName("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "invalid key", StatusCode: StatusInvalidKey})
		return
	}

	_, exists, err := s.store.Get(req.Context(), key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Message: "key not found", StatusCode: StatusKeyNotFound})
		return
	}

	if err := s.store.Delete(req.Context(), key); err != nil {
		s.log.Error().Err(err).Msg("failed to delete key")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "failed to delete key", StatusCode: StatusStorageError})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "key deleted successfully", StatusCode: StatusSuccess})
}
