package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ShortenerRouter(
	shortURLAndStore func(string) (string, error),
	getURL func(string) (string, error),
	pingDB func(ctx context.Context) error,
) http.Handler {
	r := chi.NewRouter()
	r.Use(compressionMiddleware)
	r.Use(decompressMiddleware)

	r.Get("/ping", createPingHandler(pingDB))
	r.Get("/{id}", redirectHandler(getURL))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	r.Post("/", createShortURLHandler(shortURLAndStore))
	r.Post("/api/shorten", createShortURLHandlerAPIShorten(shortURLAndStore))

	return r
}
