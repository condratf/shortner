package app

import (
	"fmt"
	"log"
	"net/url"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/db"
	"github.com/condratf/shortner/internal/app/sharedtypes"
	"github.com/condratf/shortner/internal/app/shortener"
	"github.com/condratf/shortner/internal/app/storage"
)

func shortURLAndStore(
	short shortener.Shortener,
	store storage.Storage,
) func(originalURL string) (string, error) {
	var inner func(originalURL string) (string, error)

	inner = func(originalURL string) (string, error) {
		key, err := short.Shorten(originalURL)
		if err != nil {
			return "", err
		}
		if url, _ := store.Get(key); url != "" {
			return inner(originalURL)
		}

		store.Save(key, originalURL)
		store.SaveToFile(config.Config.FilePath)

		baseURL, err := url.Parse(config.Config.BaseURL)
		if err != nil {
			return "", fmt.Errorf("invalid base URL: %w", err)
		}

		baseURL.Path = baseURL.Path + "/" + key

		return baseURL.String(), nil
	}

	return inner
}

func getURL(store storage.Storage) func(key string) (string, error) {
	return func(key string) (string, error) {
		store.LoadFromFile(config.Config.FilePath)
		url, err := store.Get(key)

		if err != nil {
			return "", err
		}
		return url, nil
	}
}

func shortURLAndStoreBatch(
	short shortener.Shortener,
	store storage.Storage,
) func(origURLs []sharedtypes.RequestPayloadBatch) ([]storage.BatchItem, error) {
	return func(origURLs []sharedtypes.RequestPayloadBatch) ([]storage.BatchItem, error) {
		var batchData []storage.BatchItem

		for _, orig := range origURLs {
			key, err := short.Shorten(orig.OriginalURL)
			if err != nil {
				return nil, fmt.Errorf("failed to shorten URL %s: %w", orig.OriginalURL, err)
			}

			if _, err = store.Get(key); err == nil {
				continue
			}

			batchData = append(batchData, storage.BatchItem{
				CorrelationID: orig.CorrelationID,
				ShortURL:      key,
				OriginalURL:   orig.OriginalURL,
			})
		}

		_, err := store.SaveBatch(batchData)
		if err != nil {
			return nil, fmt.Errorf("failed to save batch: %w", err)
		}

		return batchData, nil
	}
}

func initStore() (storage.Storage, error) {
	if config.Config.DatabaseDSN != "" {
		if err := db.InitDB(); err != nil {
			log.Printf("Failed to initialize database: %v", err)
			return nil, err
		}
		store, err := storage.NewPostgresStore(db.DB)
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
			return nil, err
		}
		return store, nil
	}

	if config.Config.FilePath != "" {
		fileStore := storage.NewInMemoryStore()
		err := fileStore.LoadFromFile(config.Config.FilePath)
		if err != nil {
			log.Fatalf("Failed to load from file: %v", err)
			return nil, err
		}
		return fileStore, nil
	}

	return storage.NewInMemoryStore(), nil
}
