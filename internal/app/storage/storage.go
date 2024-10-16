package storage

import "errors"

type Storage interface {
	Save(id, url string)
	Get(id string) (string, error)
}

type InMemoryStore struct {
	data map[string]string
}

func NewInMemoryStore() Storage {
	return &InMemoryStore{data: make(map[string]string)}
}

func (s *InMemoryStore) Save(id, url string) {
	s.data[id] = url
}

func (s *InMemoryStore) Get(id string) (string, error) {
	url, ok := s.data[id]
	if !ok {
		return "", errors.New("url not found")
	}
	return url, nil
}
