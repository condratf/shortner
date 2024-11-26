package app

import (
	"fmt"
	"net/url"

	"github.com/condratf/shortner/internal/app/config"
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
