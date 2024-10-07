package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenerRouter(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:           "GET request with valid ID",
			method:         http.MethodGet,
			path:           "/valid-id",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "http://example.com",
		},
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
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "POST request with empty body",
			method:         http.MethodPost,
			path:           "/",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "could not read request body",
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
			// reset storage
			storage = make(map[string]string)

			// populate the storage if the test involves a valid ID
			if tt.path == "/valid-id" {
				storage["valid-id"] = "http://example.com"
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			recorder := httptest.NewRecorder()

			router := ShortenerRouter()
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
