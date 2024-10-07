package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"

	"github.com/condratf/shortner/internal/app/config"
)

type urlStorage map[string]string

const (
	urlLength   = 9
	charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetSize = int64(len(charset))
)

var (
	storage = urlStorage{}
)

type URLNotExistError struct{}

func (m *URLNotExistError) Error() string {
	return "the url does not exist"
}

func generateShortedKey() (string, error) {
	shorted := make([]byte, urlLength)

	for i := range shorted {
		num, err := rand.Int(rand.Reader, big.NewInt(charsetSize))
		if err != nil {
			return "", err
		}
		shorted[i] = charset[num.Int64()]
	}

	return string(shorted), nil
}

func shortURLAndStore(originalURL string) (string, error) {
	key, err := generateShortedKey()

	if err != nil {
		return "", err
	}

	if _, ok := storage[key]; ok {
		return shortURLAndStore(originalURL)
	}

	storage[key] = originalURL

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	baseURL.Path = baseURL.Path + "/" + key

	return baseURL.String(), nil
}

func getURL(key string) (string, error) {
	if url, ok := storage[key]; ok {
		return url, nil
	}
	return "", &URLNotExistError{}
}
