package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/condratf/shortner/internal/app/errorhandler"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type requestPayload struct {
	URL string `json:"url"`
}

type responsePayload struct {
	Result string `json:"result"`
}

func createShortURLHandlerAPIShorten(shortURLAndStore models.ShortURLAndStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req requestPayload
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || len(req.URL) < 1 {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		userID, err := getUserIDFromCookie(r)
		if err != nil {
			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Println("no user id")
		}

		shortURL, err := shortURLAndStore(req.URL, userID)
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
	shortURLAndStoreBatch models.ShortURLAndStoreBatch,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []models.RequestPayloadBatch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req) == 0 {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		userID, err := getUserIDFromCookie(r)
		if err != nil {
			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Println("no user id")
		}

		batchData, err := shortURLAndStoreBatch(req, userID)
		if err != nil {
			fmt.Println(err)
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

func createShortURLHandler(shortURLAndStore models.ShortURLAndStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil || len(url) == 0 {
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromCookie(r)
		if err != nil {
			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Println("no user id")
		}

		shortURL, err := shortURLAndStore(string(url), userID)
		if err != nil {
			log.Println(err)
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

func redirectHandler(store storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		url, err := store.Get(id)

		if err != nil {
			if err.Error() == "url is deleted" {
				w.WriteHeader(http.StatusGone)
				return
			}
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

func getUserURLsHandler(store storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromCookie(r)
		if err != nil || userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		urls, err := store.GetUserURLs(*userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//return 401 if no urls
		if len(urls) == 0 {
			http.Error(w, "No URLs found", http.StatusUnauthorized)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(urls)
	}
}

func deleteURLHandler(store storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromCookie(r)
		if err != nil || userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var ids []string
		err = json.NewDecoder(r.Body).Decode(&ids)
		defer r.Body.Close()

		if err != nil || len(ids) == 0 {
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)

		go func() {
			err = store.DeleteURLs(ids, *userID)
			if err != nil {
				log.Printf("could not delete URL: %v", err)
			}
		}()
	}
}
