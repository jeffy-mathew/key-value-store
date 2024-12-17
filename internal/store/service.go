// Package store provides a service for managing a key-value store.
//
// This package includes the definition of a Service struct which
// encapsulates methods for setting, getting, and deleting keys in
// the key-value store.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"

	"codesignal/internal/repository"
)

var (
	ErrKeyTooLong    = errors.New("key length exceeds maximum allowed length")
	ErrValueTooLarge = errors.New("value size exceeds maximum allowed size")
)

// Validation constants
const (
	DefaultMaxKeyLength = 256     // Maximum length for keys in characters
	DefaultMaxValueSize = 1 << 20 // Maximum size for values (1MB)
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
	StatusSuccess       StatusCode = 1000
	StatusKeyNotFound   StatusCode = 1001
	StatusKeyExists     StatusCode = 1002
	StatusInvalidKey    StatusCode = 1003
	StatusInvalidValue  StatusCode = 1004
	StatusStorageError  StatusCode = 1005
	StatusInvalidJSON   StatusCode = 1006
	StatusKeyTooLong    StatusCode = 1007
	StatusValueTooLarge StatusCode = 1008
)

// Response represents the API response
type Response struct {
	Message    string     `json:"message"`
	StatusCode StatusCode `json:"status_code"`
	Data       KeyValue   `json:"data,omitempty"`
}

// Service for managing a key value store.
type Service struct {
	maxKeyLength int
	MaxValueSize int
	log          zerolog.Logger
	store        repository.Store
}

type Opts struct {
	MaxKeyLength int
	MaxValueSize int
}

// NewService returns a new instance of Service.
func NewService(log zerolog.Logger, store repository.Store, opts Opts) *Service {
	return &Service{
		maxKeyLength: opts.MaxKeyLength,
		MaxValueSize: opts.MaxValueSize,
		log:          log,
		store:        store,
	}
}

func (s *Service) getMaxKeyLength() int {
	if s.maxKeyLength <= 0 {
		return DefaultMaxKeyLength
	}

	return s.maxKeyLength
}

func (s *Service) getMaxValueSize() int {
	if s.MaxValueSize <= 0 {
		return DefaultMaxValueSize
	}

	return s.MaxValueSize
}

// validateKeyValue checks if the key-value pair meets the size requirements
func (s *Service) validateKeyValue(kv KeyValue) error {
	if len(kv.Key) > s.getMaxKeyLength() {
		return fmt.Errorf("err: %w, max key length: %d", ErrKeyTooLong, s.getMaxKeyLength())
	}
	if len(kv.Value) > s.getMaxValueSize() {
		return fmt.Errorf("err: %w, max value size: %d", ErrValueTooLarge, s.getMaxValueSize())
	}
	return nil
}

func (s *Service) SetKey(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		s.log.Error().Err(err).Msg("failed to decode request body")
		s.doJSONWrite(w, http.StatusBadRequest, Response{Message: "invalid request body", StatusCode: StatusInvalidJSON})
		return
	}

	if err := s.validateKeyValue(kv); err != nil {
		s.log.Error().Err(err).Msg("invalid key-value pair")
		var statusCode StatusCode
		if errors.Is(err, ErrKeyTooLong) {
			statusCode = StatusKeyTooLong
		} else {
			statusCode = StatusValueTooLarge
		}
		s.doJSONWrite(w, http.StatusBadRequest, Response{Message: err.Error(), StatusCode: statusCode})
		return
	}

	_, exists, err := s.store.Get(r.Context(), kv.Key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		s.doJSONWrite(w, http.StatusInternalServerError, Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if exists {
		s.doJSONWrite(w, http.StatusConflict, Response{Message: "key already exists", StatusCode: StatusKeyExists})
		return
	}

	if err := s.store.Set(r.Context(), kv.Key, []byte(kv.Value)); err != nil {
		s.log.Error().Err(err).Msg("failed to set key")
		s.doJSONWrite(w, http.StatusInternalServerError, Response{Message: "failed to set key", StatusCode: StatusStorageError})
		return
	}

	s.doJSONWrite(w, http.StatusCreated, Response{Message: "key created successfully", StatusCode: StatusSuccess})
}

func (s *Service) GetKey(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	key := params.ByName("key")
	if key == "" {
		s.doJSONWrite(w, http.StatusBadRequest, Response{Message: "invalid key", StatusCode: StatusInvalidKey})
		return
	}

	kv, exists, err := s.store.Get(r.Context(), key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		s.doJSONWrite(w, http.StatusInternalServerError, Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if !exists {
		s.doJSONWrite(w, http.StatusNotFound, Response{Message: "key not found", StatusCode: StatusKeyNotFound})
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
		s.doJSONWrite(w, http.StatusBadRequest, Response{Message: "invalid key", StatusCode: StatusInvalidKey})
		return
	}

	_, exists, err := s.store.Get(req.Context(), key)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get key")
		s.doJSONWrite(w, http.StatusInternalServerError, Response{Message: "failed to get key", StatusCode: StatusStorageError})
		return
	}

	if !exists {
		s.doJSONWrite(w, http.StatusNotFound, Response{Message: "key not found", StatusCode: StatusKeyNotFound})
		return
	}

	if err := s.store.Delete(req.Context(), key); err != nil {
		s.log.Error().Err(err).Msg("failed to delete key")
		s.doJSONWrite(w, http.StatusInternalServerError, Response{Message: "failed to delete key", StatusCode: StatusStorageError})
		return
	}

	s.doJSONWrite(w, http.StatusOK, Response{Message: "key deleted successfully", StatusCode: StatusSuccess})
}

func (s *Service) doJSONWrite(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		s.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
