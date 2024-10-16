package shortener

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	urlLength   = 9
	charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetSize = int64(len(charset))
)

type Shortener interface {
	Shorten(url string) (string, error)
}

type DefaultShortener struct{}

func NewShortener() Shortener {
	return &DefaultShortener{}
}

func (s *DefaultShortener) Shorten(originalURL string) (string, error) {
	var builder strings.Builder
	builder.Grow(urlLength)

	for i := 0; i < urlLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(charsetSize))
		if err != nil {
			return "", err
		}
		builder.WriteByte(charset[num.Int64()])
	}

	return builder.String(), nil
}
