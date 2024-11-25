package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UUID = string

type BatchItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
}

type Storage interface {
	Save(shortURL, originalURL string) (UUID, error)
	SaveBatch([]BatchItem) ([]URLData, error)
	Get(id string) (string, error)
	LoadFromFile(filePath string) error
	SaveToFile(filePath string) error
}

type InMemoryStore struct {
	data map[string]URLData
}

type ErrURLExists struct {
	ID               string
	ExistingShortURL string
}

func (e *ErrURLExists) Error() string {
	return fmt.Sprintf("url already exists with short URL: %s", e.ExistingShortURL)
}

func (e *ErrURLExists) Is(target error) bool {
	_, ok := target.(*ErrURLExists)
	return ok
}

func NewInMemoryStore() Storage {
	return &InMemoryStore{data: make(map[string]URLData)}
}

func (s *InMemoryStore) Save(shortURL, originalURL string) (string, error) {
	id := uuid.New().String()
	s.data[shortURL] = URLData{
		UUID:        id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return id, nil
}

func (s *InMemoryStore) SaveBatch(items []BatchItem) ([]URLData, error) {
	var urlDataList []URLData
	for _, item := range items {
		urlDataList = append(urlDataList, URLData{
			UUID:        item.CorrelationID,
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
		})
		s.data[item.ShortURL] = URLData{
			UUID:        item.CorrelationID,
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
		}
	}
	return urlDataList, nil
}

func (s *InMemoryStore) Get(shortURL string) (string, error) {
	urlData, ok := s.data[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	return urlData.OriginalURL, nil
}

func (s *InMemoryStore) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var urlDataList []URLData
	if err := json.NewDecoder(file).Decode(&urlDataList); err != nil {
		return err
	}

	for _, urlData := range urlDataList {
		s.data[urlData.ShortURL] = urlData
	}

	return nil
}

func (s *InMemoryStore) SaveToFile(filePath string) error {
	var urlDataList []URLData
	for _, urlData := range s.data {
		urlDataList = append(urlDataList, urlData)
	}

	data, err := json.Marshal(urlDataList)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
