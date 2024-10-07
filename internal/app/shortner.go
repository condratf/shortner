package app

import (
	"crypto/rand"
	"math/big"
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
	shortURL := make([]byte, urlLength)

	for i := range shortURL {
		num, err := rand.Int(rand.Reader, big.NewInt(charsetSize))
		if err != nil {
			return "", err
		}
		shortURL[i] = charset[num.Int64()]
	}

	return string(shortURL), nil
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
