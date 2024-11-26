package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/db"
	"github.com/condratf/shortner/internal/app/logger"
	"github.com/condratf/shortner/internal/app/router"
	"github.com/condratf/shortner/internal/app/shortener"

	"github.com/go-chi/chi/v5"
)

func Server() error {
	config.InitConfig()
	short := shortener.NewShortener()
	store, err := initStore()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
		return err
	}
	if db.DB != nil {
		defer db.CloseDB()
	}

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware())

	shortenerRouter := router.ShortenerRouter(
		shortURLAndStore(short, store),
		getURL(store),
		shortURLAndStoreBatch(short, store),
		db.PingDB,
	)
	r.Mount("/", shortenerRouter)
	fmt.Printf("starting server at :%s\n", config.Config.Addr)

	return http.ListenAndServe(config.Config.Addr, r)
}
