package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Invalid method (GET request)",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "POST request with valid URL",
			method:         http.MethodPost,
			body:           "http://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
		{
			name:           "POST request with empty body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset storage for each test
			storage = make(map[string]string)

			req := httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString(tt.body))
			recorder := httptest.NewRecorder()

			postHandler(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusCreated {
				responseBody, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(responseBody), "http://localhost:8080/")
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Invalid method (POST request)",
			method:         http.MethodPost,
			path:           "/valid-id",
			expectedStatus: http.StatusMethodNotAllowed,
		},
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
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset storage for each test
			storage = make(map[string]string)

			// populate the storage
			if tt.path == "/valid-id" {
				storage["valid-id"] = "http://example.com"
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			getHandler(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				location := recorder.Header().Get("location")
				assert.Equal(t, tt.expectedHeader, location)
			}
		})
	}

}
