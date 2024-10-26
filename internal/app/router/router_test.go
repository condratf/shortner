package router

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestShortenerRouter(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		body             string
		expectedStatus   int
		expectedBody     string
		expectedHeader   string
		shortURLAndStore func(string) (string, error)
		getURL           func(string) (string, error)
	}{
		{
			name:           "GET request with valid ID",
			method:         http.MethodGet,
			path:           "/valid-id",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "http://example.com",
			getURL: func(id string) (string, error) {
				if id == "valid-id" {
					return "http://example.com", nil
				}
				return "", errors.New("invalid ID")
			},
		},
		{
			name:           "GET request with invalid ID",
			method:         http.MethodGet,
			path:           "/invalid-id",
			expectedStatus: http.StatusBadRequest,
			getURL: func(id string) (string, error) {
				return "", errors.New("invalid ID")
			},
		},
		{
			name:           "GET request with no ID",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusNotFound,
			getURL:         nil,
		},
		{
			name:           "POST request with valid URL",
			method:         http.MethodPost,
			path:           "/",
			body:           "http://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   config.Config.BaseURL,
			shortURLAndStore: func(url string) (string, error) {
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
			shortURLAndStore: func(url string) (string, error) {
				return "", nil
			},
		},
		{
			name:           "Invalid method (PUT request)",
			method:         http.MethodPut,
			path:           "/valid-id",
			expectedStatus: http.StatusMethodNotAllowed,
			getURL:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			recorder := httptest.NewRecorder()

			router := ShortenerRouter(tt.shortURLAndStore, tt.getURL)
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
