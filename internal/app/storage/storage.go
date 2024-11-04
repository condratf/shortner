package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UUID = string

type Storage interface {
	Save(shortURL, originalURL string) (UUID, error)
	Get(id string) (string, error)
	LoadFromFile(filePath string) error
	SaveToFile(filePath string) error
}

type InMemoryStore struct {
	data map[string]URLData
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