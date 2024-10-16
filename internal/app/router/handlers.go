package router

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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
