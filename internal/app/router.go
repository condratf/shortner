package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ShortenerRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/{id}", getHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	r.Post("/", postHandler)

	return r
}
