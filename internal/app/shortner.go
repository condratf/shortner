package app

import (
	"crypto/rand"
	"encoding/base64"
)

type urlStorage map[string]string

const (
	urlLength = 9
)

var (
	storage = urlStorage{}
)

type URLNotExistError struct{}

func (m *URLNotExistError) Error() string {
	return "the url does not exist"
}

func generateShortedKey() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	shortURL := base64.URLEncoding.EncodeToString(bytes)[:urlLength]
	return shortURL, nil
}

func shortURLAndStore(url string) (string, error) {
	key, err := generateShortedKey()

	if err != nil {
		return "", err
	}

	if _, ok := storage[key]; ok {
		return shortURLAndStore(url)
	}

	storage[key] = url

	return "http://localhost:8080/" + key, nil
}

func getURL(key string) (string, error) {
	if url, ok := storage[key]; ok {
		return url, nil
	}
	return "", &URLNotExistError{}
}
