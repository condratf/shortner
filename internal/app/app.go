package app

import (
	"fmt"
	"net/http"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/logger"
	"github.com/condratf/shortner/internal/app/router"
	"github.com/condratf/shortner/internal/app/shortener"
	"github.com/condratf/shortner/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func Server() error {
	config.InitConfig()
	short := shortener.NewShortener()
	store := storage.NewInMemoryStore()
	r := chi.NewRouter()

	r.Use(logger.LoggingMiddleware(logger.InitLogger()))

	shortenerRouter := router.ShortenerRouter(shortURLAndStore(short, store), getURL(store))
	r.Mount("/", shortenerRouter)
	fmt.Printf("starting server at :%s\n", config.Config.Addr)

	return http.ListenAndServe(config.Config.Addr, r)
}
