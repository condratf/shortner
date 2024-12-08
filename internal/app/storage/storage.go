package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/condratf/shortner/internal/app/models"
	"github.com/google/uuid"
)

const (
	FilePermUserReadWrite = 0600
	FilePermUserGroupRead = 0640
	FilePermAllReadWrite  = 0666
	FilePermAllReadOnly   = 0644
)

type URLData struct {
	UUID        string `json:"uuid" db:"user_id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	UserID      string `json:"user_id" db:"user_id"`
	DeletedFlag bool   `json:"deleted_flag" db:"deleted_flag"`
}

type UUID = string

type Storage interface {
	Save(shortURL, originalURL string, userID *string) (UUID, error)
	SaveBatch(items []models.BatchItem, userID *string) ([]URLData, error)
	Get(id string) (string, error)
	GetUserURLs(userID string) ([]models.UserURLs, error)
	LoadFromFile(filePath string) error
	SaveToFile(filePath string) error
	DeleteURLs(shortURLs []string, userID string) error
}

type InMemoryStore struct {
	data map[string]URLData
	mu   sync.RWMutex
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

func (s *InMemoryStore) Save(shortURL, originalURL string, userID *string) (string, error) {
	id := uuid.New().String()
	s.mu.Lock()
	defer s.mu.Unlock()

	urlData := URLData{
		UUID:        id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	if userID != nil {
		urlData.UserID = *userID
	}

	s.data[shortURL] = urlData
	return id, nil
}

func (s *InMemoryStore) SaveBatch(items []models.BatchItem, userID *string) ([]URLData, error) {
	var urlDataList []URLData
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		urlData := URLData{
			UUID:        item.CorrelationID,
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
		}

		if userID != nil {
			urlData.UserID = *userID
		}

		urlDataList = append(urlDataList, urlData)
		s.data[item.ShortURL] = urlData
	}
	return urlDataList, nil
}

func (s *InMemoryStore) Get(shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlData, ok := s.data[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}
	if urlData.DeletedFlag {
		return "", errors.New("url is deleted")
	}
	return urlData.OriginalURL, nil
}

func (s *InMemoryStore) GetUserURLs(userID string) ([]models.UserURLs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userURLs []models.UserURLs
	for _, urlData := range s.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, models.UserURLs{
				ShortURL:    urlData.ShortURL,
				OriginalURL: urlData.OriginalURL,
			})
		}
	}
	return userURLs, nil
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

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, urlData := range urlDataList {
		s.data[urlData.ShortURL] = urlData
	}

	return nil
}

func (s *InMemoryStore) SaveToFile(filePath string) error {
	s.mu.RLock()
	var urlDataList []URLData
	for _, urlData := range s.data {
		urlDataList = append(urlDataList, urlData)
	}
	s.mu.RUnlock()

	data, err := json.Marshal(urlDataList)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FilePermAllReadOnly)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (s *InMemoryStore) DeleteURLs(shortURLs []string, userID string) error {
	var wg sync.WaitGroup
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shortURL := range shortURLs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			s.mu.Lock()
			defer s.mu.Unlock()
			if urlData, exists := s.data[id]; exists && urlData.UserID == userID {
				urlData.DeletedFlag = true
				s.data[id] = urlData
			}
		}(shortURL)
	}

	wg.Wait()
	return nil
}
