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
	"github.com/condratf/shortner/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func Server() error {
	config.InitConfig()
	short := shortener.NewShortener()
	var store storage.Storage

	// Determine storage based on priority
	if config.Config.DatabaseDSN != "" {
		if err := db.InitDB(); err != nil {
			log.Printf("Failed to initialize database: %v", err)
		} else {
			defer db.CloseDB()
			store, err = storage.NewPostgresStore(db.DB)
			if err != nil {
				log.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
			}
		}
	} else if config.Config.FilePath != "" {
		fileStore := storage.NewInMemoryStore()
		err := fileStore.LoadFromFile(config.Config.FilePath)
		if err != nil {
			log.Fatalf("Failed to load from file: %v", err)
		}
		store = fileStore
	} else {
		store = storage.NewInMemoryStore()
	}

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware(logger.InitLogger()))

	shortenerRouter := router.ShortenerRouter(shortURLAndStore(short, store), getURL(store), db.PingDB)
	r.Mount("/", shortenerRouter)
	fmt.Printf("starting server at :%s\n", config.Config.Addr)

	return http.ListenAndServe(config.Config.Addr, r)
}
