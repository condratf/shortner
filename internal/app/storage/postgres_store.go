package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) (Storage, error) {
	store := &PostgresStore{db: db}

	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id UUID PRIMARY KEY,
			short_url TEXT UNIQUE NOT NULL,
			original_url TEXT NOT NULL
		)
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) Save(shortURL, originalURL string) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO urls (id, short_url, original_url) VALUES ($1, $2, $3)`

	_, err := s.db.Exec(query, id, shortURL, originalURL)
	if err != nil {
		return "", fmt.Errorf("could not insert url: %w", err)
	}

	return id, nil
}

func (s *PostgresStore) Get(shortURL string) (string, error) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1`

	err := s.db.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("url not found")
		}
		return "", fmt.Errorf("could not get url: %w", err)
	}

	return originalURL, nil
}

func (s *PostgresStore) LoadFromFile(_ string) error {
	// не поддерживаем загрузку из файла
	return nil
}

func (s *PostgresStore) SaveToFile(_ string) error {
	// не поддерживаем сохранение в файл
	return nil
}
