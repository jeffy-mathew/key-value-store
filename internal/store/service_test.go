package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	repomock "codesignal/internal/repository/mock"
	"codesignal/internal/store"
)

const (
	testKey = "test-key"
	testValue = "test-value"
)

func setupTest(t *testing.T) (*store.Service, *repomock.MockStore) {
	ctrl := gomock.NewController(t)
	mockStore := repomock.NewMockStore(ctrl)
	logger := zerolog.New(nil)
	service := store.NewService(logger, mockStore)
	return service, mockStore
}

func TestServiceSet(t *testing.T) {
	tests := []struct {
		name           string
		input          store.KeyValue
		setupMock      func(*repomock.MockStore)
		expectedStatus int
		expectedBody   store.Response
	}{
		{
			name: "key fetch failed due to storage error",
			input: store.KeyValue{
				Key:   "existing-key",
				Value: testValue,
			},
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), "existing-key").
					Return(nil, false, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: store.Response{
				Message:    "failed to get key",
				StatusCode: store.StatusStorageError,
			},
		},
		{
			name: "key already exists",
			input: store.KeyValue{
				Key:   "existing-key",
				Value: testValue,
			},
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), "existing-key").
					Return([]byte("existing-value"), true, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: store.Response{
				Message:    "key already exists",
				StatusCode: store.StatusKeyExists,
			},
		},
		{
			name: "key storage failed",
			input: store.KeyValue{
				Key:   testKey,
				Value: testValue,
			},
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return(nil, false, nil)
				m.EXPECT().
					Set(gomock.Any(), testKey, []byte(testValue)).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: store.Response{
				Message:    "failed to set key",
				StatusCode: store.StatusStorageError,
			},
		},
		{
			name: "success",
			input: store.KeyValue{
				Key:   testKey,
				Value: testValue,
			},
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return(nil, false, nil)
				m.EXPECT().
					Set(gomock.Any(), testKey, []byte(testValue)).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: store.Response{
				Message:    "key created successfully",
				StatusCode: store.StatusSuccess,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockStore := setupTest(t)
			tt.setupMock(mockStore)

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/key", bytes.NewReader(body))
			w := httptest.NewRecorder()

			service.SetKey(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response store.Response
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestServiceGet(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*repomock.MockStore)
		expectedStatus int
		expectedBody   store.Response
	}{
		{
			name:           "empty key",
			key:            "",
			setupMock:      func(m *repomock.MockStore) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: store.Response{
				Message:    "invalid key",
				StatusCode: store.StatusInvalidKey,
			},
		},
		{
			name: "key fetch error",
			key:  testKey,
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return(nil, false, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: store.Response{
				Message:    "failed to get key",
				StatusCode: store.StatusStorageError,
			},
		},
		{
			name: "key not found",
			key:  "non-existent-key",
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), "non-existent-key").
					Return(nil, false, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: store.Response{
				Message:    "key not found",
				StatusCode: store.StatusKeyNotFound,
			},
		},
		{
			name: "success",
			key:  testKey,
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return([]byte(testValue), true, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: store.Response{
				Message:    "key found successfully",
				StatusCode: store.StatusSuccess,
				Data: store.KeyValue{
					Key:   testKey,
					Value: testValue,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockStore := setupTest(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/key/%s", tt.key), nil)
			w := httptest.NewRecorder()
			params := httprouter.Params{{Key: "key", Value: tt.key}}
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

			service.GetKey(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response store.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*repomock.MockStore)
		expectedStatus int
		expectedBody   store.Response
	}{
		{
			name: "key storage failed",
			key:  testKey,
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return(nil, false, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: store.Response{
				Message:    "failed to get key",
				StatusCode: store.StatusStorageError,
			},
		},
		{
			name: "key not found",
			key:  "non-existent-key",
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), "non-existent-key").
					Return(nil, false, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: store.Response{
				Message:    "key not found",
				StatusCode: store.StatusKeyNotFound,
			},
		},
		{
			name: "success",
			key:  testKey,
			setupMock: func(m *repomock.MockStore) {
				m.EXPECT().
					Get(gomock.Any(), testKey).
					Return([]byte(testValue), true, nil)
				m.EXPECT().
					Delete(gomock.Any(), testKey).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: store.Response{
				Message:    "key deleted successfully",
				StatusCode: store.StatusSuccess,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockStore := setupTest(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest(http.MethodDelete, "/v1/store/"+tt.key, nil)
			w := httptest.NewRecorder()
			params := httprouter.Params{{Key: "key", Value: tt.key}}
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

			service.DeleteKey(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response store.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}
