package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/sharedtypes"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/condratf/shortner/internal/app/utils"
	"github.com/go-chi/chi/v5"
)

type requestPayload struct {
	URL string `json:"url"`
}

type responsePayload struct {
	Result string `json:"result"`
}

func handleURLExistError(w http.ResponseWriter, err error, respType string) bool {
	if errors.Is(err, &storage.ErrURLExists{}) {
		var urlExistsErr *storage.ErrURLExists
		if errors.As(err, &urlExistsErr) {
			shortURL, err := utils.ConstructURL(config.Config.BaseURL, urlExistsErr.ExistingShortURL)
			fmt.Println(shortURL)
			if err != nil {
				http.Error(w, "could not construct URL", http.StatusInternalServerError)
				return true
			}
			shortURL = strings.TrimSpace(shortURL)
			switch respType {
			case "json":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				if err := json.NewEncoder(w).Encode(responsePayload{Result: shortURL}); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
			case "json-batch":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				if err := json.NewEncoder(w).Encode(sharedtypes.ResponsePayloadBatch{
					CorrelationID: urlExistsErr.ID,
					ShortURL:      shortURL,
				}); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
			case "text":
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusConflict)
				_, writeErr := w.Write([]byte(shortURL))
				if writeErr != nil {
					http.Error(w, "could not write response", http.StatusInternalServerError)
				}
			}
			return true
		}
	}
	return false
}

func createShortURLHandlerAPIShorten(shortURLAndStore func(string) (string, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req requestPayload
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || len(req.URL) < 1 {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		shortURL, err := shortURLAndStore(req.URL)
		if err != nil {
			if handleURLExistError(w, err, "json") {
				return
			}
			http.Error(w, "could not store URL", http.StatusInternalServerError)
			return
		}

		resp := responsePayload{Result: shortURL}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "could not encode response", http.StatusInternalServerError)
		}
	}
}

func createShortURLHandlerAPIShortenBatch(
	shortURLAndStoreBatch func([]sharedtypes.RequestPayloadBatch) ([]storage.BatchItem, error),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []sharedtypes.RequestPayloadBatch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req) == 0 {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		batchData, err := shortURLAndStoreBatch(req)
		if err != nil {
			if handleURLExistError(w, err, "json-batch") {
				return
			}
			http.Error(w, "Failed to process batch", http.StatusInternalServerError)
			return
		}

		resp := make([]sharedtypes.ResponsePayloadBatch, len(batchData))
		for i, item := range batchData {
			resp[i] = sharedtypes.ResponsePayloadBatch{
				CorrelationID: item.CorrelationID,
				ShortURL:      item.ShortURL,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func createShortURLHandler(shortURLAndStore func(string) (string, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil || len(url) == 0 {
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}

		shortURL, err := shortURLAndStore(string(url))
		if err != nil {
			if handleURLExistError(w, err, "text") {
				return
			}
			http.Error(w, "could not store URL", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func redirectHandler(getURL func(string) (string, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		url, err := getURL(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func createPingHandler(pingDB func(ctx context.Context) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := pingDB(ctx); err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
