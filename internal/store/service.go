// Package store provides a service for managing a key-value store.
//
// This package includes the definition of a Service struct which
// encapsulates methods for setting, getting, and deleting keys in
// the key-value store.
//
// (!) Task: Implement the SetKey, GetKey, and DeleteKey methods on the Service.
package store

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Service for managing a key value store.
type Service struct {
	log zerolog.Logger
}

// NewService returns a new instance of Service.
func NewService(log zerolog.Logger) *Service {
	return &Service{
		log: log,
	}
}

func (s *Service) SetKey(w http.ResponseWriter, req *http.Request) {
	// (!) Implement me.
	w.WriteHeader(http.StatusOK)
}

func (s *Service) GetKey(w http.ResponseWriter, req *http.Request) {
	// (!) Implement me.
	w.WriteHeader(http.StatusOK)
}

func (s *Service) DeleteKey(w http.ResponseWriter, req *http.Request) {
	// (!) Implement me.
	w.WriteHeader(http.StatusOK)
}
