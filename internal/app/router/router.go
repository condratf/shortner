package router

import (
	"context"
	"net/http"

	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func ShortenerRouter(
	shortURLAndStore models.ShortURLAndStore,
	shortURLAndStoreBatch models.ShortURLAndStoreBatch,
	getURL func(string) (string, error),
	pingDB func(ctx context.Context) error,
	store storage.Storage,
) http.Handler {
	r := chi.NewRouter()
	r.Use(compressionMiddleware)
	r.Use(decompressMiddleware)
	r.Use(userAuthMiddleware)

	r.Get("/ping", createPingHandler(pingDB))
	r.Get("/api/user/urls", getUserURLsHandler(store))
	r.Get("/{id}", redirectHandler(getURL))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	r.Post("/", createShortURLHandler(shortURLAndStore))
	r.Post("/api/shorten", createShortURLHandlerAPIShorten(shortURLAndStore))
	r.Post("/api/shorten/batch", createShortURLHandlerAPIShortenBatch(shortURLAndStoreBatch))

	return r
}
