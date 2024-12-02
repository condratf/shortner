package app

import (
	"errors"
	"fmt"
	"log"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/db"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/shortener"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/condratf/shortner/internal/app/utils"
)

func shortURLAndStore(
	short shortener.Shortener,
	store storage.Storage,
) models.ShortURLAndStore {
	var inner models.ShortURLAndStore

	inner = func(originalURL string, userID *string) (string, error) {
		key, err := short.Shorten(originalURL)
		if err != nil {
			return "", err
		}
		if url, _ := store.Get(key); url != "" {
			return inner(originalURL, userID)
		}

		_, err = store.Save(key, originalURL, userID)
		if errors.Is(err, &storage.ErrURLExists{}) {
			fmt.Println("URL already exists")
			return "", err
		}
		store.SaveToFile(config.Config.FilePath)

		shortURL, err := utils.ConstructURL(config.Config.BaseURL, key)
		if err != nil {
			return "", err
		}

		return shortURL, nil
	}

	return inner
}

func getURL(store storage.Storage) func(key string) (string, error) {
	return func(key string) (string, error) {
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
) models.ShortURLAndStoreBatch {
	return func(origURLs []models.RequestPayloadBatch, userID *string) ([]models.BatchItem, error) {
		var batchData []models.BatchItem
		var batchDataResponse []models.BatchItem

		for _, orig := range origURLs {
			key, err := short.Shorten(orig.OriginalURL)
			if err != nil {
				return nil, fmt.Errorf("failed to shorten URL %s: %w", orig.OriginalURL, err)
			}

			batchData = append(batchData, models.BatchItem{
				CorrelationID: orig.CorrelationID,
				ShortURL:      key,
				OriginalURL:   orig.OriginalURL,
			})

			shortURL, err := utils.ConstructURL(config.Config.BaseURL, key)
			if err != nil {
				return nil, err
			}

			batchDataResponse = append(batchDataResponse, models.BatchItem{
				CorrelationID: orig.CorrelationID,
				ShortURL:      shortURL,
				OriginalURL:   orig.OriginalURL,
			})
		}

		_, err := store.SaveBatch(batchData, userID)
		if err != nil {
			if errors.Is(err, &storage.ErrURLExists{}) {
				return nil, err
			}
			return nil, fmt.Errorf("failed to save batch: %w", err)
		}

		return batchDataResponse, nil
	}
}

func initStore() (storage.Storage, error) {
	if config.Config.DatabaseDSN != "" {
		if err := db.InitDB(); err != nil {
			log.Printf("Failed to initialize database: %v", err)
			return nil, err
		}

		// if err := db.ApplyMigrations(config.Config.DatabaseDSN); err != nil {
		// 	log.Printf("Failed to apply migrations: %v", err)
		// 	return nil, err
		// }

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
