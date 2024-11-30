package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/condratf/shortner/internal/app/errorhandler"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/go-chi/chi/v5"
)

type requestPayload struct {
	URL string `json:"url"`
}

type responsePayload struct {
	Result string `json:"result"`
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
			if errorhandler.HandleURLExistError(w, err, "json") {
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
	shortURLAndStoreBatch func([]models.RequestPayloadBatch) ([]models.BatchItem, error),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []models.RequestPayloadBatch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req) == 0 {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		batchData, err := shortURLAndStoreBatch(req)
		if err != nil {
			if errorhandler.HandleURLExistError(w, err, "json-batch") {
				return
			}
			http.Error(w, "Failed to process batch", http.StatusInternalServerError)
			return
		}

		resp := make([]models.ResponsePayloadBatch, len(batchData))
		for i, item := range batchData {
			resp[i] = models.ResponsePayloadBatch{
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
			if errorhandler.HandleURLExistError(w, err, "text") {
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
