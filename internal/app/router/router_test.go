package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestShortenerRouter(t *testing.T) {
	tests := []struct {
		name                  string
		method                string
		path                  string
		body                  interface{}
		expectedStatus        int
		expectedBody          string
		expectedHeader        string
		shortURLAndStore      models.ShortURLAndStore
		shortURLAndStoreBatch models.ShortURLAndStoreBatch
	}{
		{
			name:           "GET request with invalid ID",
			method:         http.MethodGet,
			path:           "/invalid-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET request with no ID",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "POST request with valid URL",
			method:         http.MethodPost,
			path:           "/",
			body:           "http://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   config.Config.BaseURL,
			shortURLAndStore: func(url string, userID *string) (string, error) {
				if url == "http://example.com" {
					return config.Config.BaseURL, nil
				}
				return "", errors.New("could not store URL")
			},
		},
		{
			name:           "POST request with empty body",
			method:         http.MethodPost,
			path:           "/",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "could not read request body",
			shortURLAndStore: func(url string, userID *string) (string, error) {
				return "", nil
			},
		},
		{
			name:           "POST request with JSON body",
			method:         http.MethodPost,
			path:           "/api/shorten",
			body:           map[string]string{"url": "http://example.com"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"` + config.Config.BaseURL + `"}`,
			shortURLAndStore: func(url string, userID *string) (string, error) {
				if url == "http://example.com" {
					return config.Config.BaseURL, nil
				}
				return "", errors.New("could not store URL")
			},
		},
		{
			name:           "POST request with empty URL in JSON",
			method:         http.MethodPost,
			path:           "/api/shorten",
			body:           map[string]string{"url": ""},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "could not decode request body",
			shortURLAndStore: func(url string, userID *string) (string, error) {
				return "", nil
			},
		},
		{
			name:           "Invalid method (PUT request)",
			method:         http.MethodPut,
			path:           "/valid-id",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				switch v := tt.body.(type) {
				case string:
					reqBody = bytes.NewBufferString(v)
				case map[string]string:
					jsonData, err := json.Marshal(v)
					assert.NoError(t, err)
					reqBody = bytes.NewBuffer(jsonData)
				}
			}

			pingDB := func(ctx context.Context) error { return nil }

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			recorder := httptest.NewRecorder()

			router := ShortenerRouter(tt.shortURLAndStore, tt.shortURLAndStoreBatch, pingDB, storage.NewInMemoryStore())
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != "" {
				responseBody, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(responseBody), tt.expectedBody)
			}

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				location := recorder.Header().Get("Location")
				assert.Equal(t, tt.expectedHeader, location)
			}
		})
	}
}
