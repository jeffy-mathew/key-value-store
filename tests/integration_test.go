//go:build integration
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"codesignal/internal/config"
	"codesignal/internal/repository"
	"codesignal/internal/router"
	"codesignal/internal/server"
	"codesignal/internal/store"
)

// IntegrationTestSuite tests the integration between different components
// of the key-value store system, including the HTTP router, store service,
// and repository layer.
type IntegrationTestSuite struct {
	suite.Suite
	srv      *httptest.Server
	client   *http.Client
	store    repository.Store
	dataFile string
}

func (s *IntegrationTestSuite) SetupSuite() {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	s.dataFile = "test_data.json"

	cfg := &config.Config{
		Server: server.Config{
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
		MaxKeyLength: 100,
		MaxValueSize: 1024,
		SyncInterval: time.Minute,
		DataFile:     s.dataFile,
	}

	// Initialize a test store
	store, err := repository.NewKeyValueStore(logger)
	s.NoError(err)
	s.store = store

	// Initialize router with dependencies
	r := router.New(logger, store, cfg)

	// Create test server
	s.srv = httptest.NewServer(r)
	s.client = &http.Client{}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.srv.Close()

	// Clean up test data file
	os.Remove(s.dataFile)
}

func (s *IntegrationTestSuite) SetupTest() {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	store, err := repository.NewKeyValueStore(logger)
	s.NoError(err)
	s.store = store

	// Reinitialize router
	r := router.New(logger, store, &config.Config{
		Server: server.Config{
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
		MaxKeyLength: 100,
		MaxValueSize: 1024,
		SyncInterval: time.Minute,
		DataFile:     s.dataFile,
	})

	s.srv.Config.Handler = r
}

func (s *IntegrationTestSuite) TestSetKeyValue() {
	testCases := []struct {
		name           string
		key            string
		value          string
		expectedCode   int
		expectedStatus store.StatusCode
	}{
		{
			name:           "Valid key-value pair",
			key:            "testkey",
			value:          "testvalue",
			expectedCode:   http.StatusCreated,
			expectedStatus: store.StatusSuccess,
		},
		{
			name:           "Empty key",
			key:            "",
			value:          "testvalue",
			expectedCode:   http.StatusCreated,
			expectedStatus: store.StatusSuccess,
		},
		{
			name:           "Long key",
			key:            strings.Repeat("x", 101), // Exceeds MaxKeyLength
			value:          "testvalue",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: store.StatusKeyTooLong,
		},
		{
			name:           "Long value",
			key:            "testkey",
			value:          strings.Repeat("x", 1025), // Exceeds MaxValueSize
			expectedCode:   http.StatusBadRequest,
			expectedStatus: store.StatusValueTooLarge,
		},
		{
			name:           "Duplicate key",
			key:            "duplicate",
			value:          "value1",
			expectedCode:   http.StatusConflict,
			expectedStatus: store.StatusKeyExists,
		},
	}

	// First set up the duplicate key test
	kv := store.KeyValue{
		Key:   "duplicate",
		Value: "initial",
	}
	jsonData, err := json.Marshal(kv)
	s.NoError(err)
	req, err := http.NewRequest(http.MethodPost, s.srv.URL+"/key", bytes.NewBuffer(jsonData))
	s.NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	s.NoError(err)
	resp.Body.Close()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			kv := store.KeyValue{
				Key:   tc.key,
				Value: tc.value,
			}
			jsonData, err := json.Marshal(kv)
			s.NoError(err)

			req, err := http.NewRequest(http.MethodPost, s.srv.URL+"/key", bytes.NewBuffer(jsonData))
			s.NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.client.Do(req)
			s.NoError(err)
			defer resp.Body.Close()

			s.Equal(tc.expectedCode, resp.StatusCode)

			var result store.Response
			err = json.NewDecoder(resp.Body).Decode(&result)
			s.NoError(err)
			s.Equal(tc.expectedStatus, result.StatusCode)
		})
	}
}

func (s *IntegrationTestSuite) TestGetKeyValue() {
	testCases := []struct {
		name           string
		setupKey       string
		setupValue     string
		getKey         string
		expectedCode   int
		expectedStatus store.StatusCode
		expectedValue  string
	}{
		{
			name:           "Existing key",
			setupKey:       "testkey",
			setupValue:     "testvalue",
			getKey:         "testkey",
			expectedCode:   http.StatusOK,
			expectedStatus: store.StatusSuccess,
			expectedValue:  "testvalue",
		},
		{
			name:           "Non-existent key",
			setupKey:       "",
			setupValue:     "",
			getKey:         "nonexistent",
			expectedCode:   http.StatusNotFound,
			expectedStatus: store.StatusKeyNotFound,
			expectedValue:  "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup if needed
			if tc.setupKey != "" {
				kv := store.KeyValue{
					Key:   tc.setupKey,
					Value: tc.setupValue,
				}
				jsonData, err := json.Marshal(kv)
				s.NoError(err)

				req, err := http.NewRequest(http.MethodPost, s.srv.URL+"/key", bytes.NewBuffer(jsonData))
				s.NoError(err)
				req.Header.Set("Content-Type", "application/json")

				resp, err := s.client.Do(req)
				s.NoError(err)
				resp.Body.Close()
				s.Equal(http.StatusCreated, resp.StatusCode)
			}

			// Test get
			req, err := http.NewRequest(http.MethodGet, s.srv.URL+"/key/"+tc.getKey, nil)
			s.NoError(err)

			resp, err := s.client.Do(req)
			s.NoError(err)
			defer resp.Body.Close()

			s.Equal(tc.expectedCode, resp.StatusCode)

			var result store.Response
			err = json.NewDecoder(resp.Body).Decode(&result)
			s.NoError(err)
			s.Equal(tc.expectedStatus, result.StatusCode)

			if tc.expectedValue != "" {
				s.Equal(tc.expectedValue, result.Data.Value)
			}
		})
	}
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
