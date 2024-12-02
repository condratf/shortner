package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/utils"
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
			original_url TEXT UNIQUE NOT NULL,
			user_id UUID
		)
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) Save(shortURL, originalURL string, userID *string) (string, error) {
	id := uuid.New().String()
	query := `
		INSERT INTO urls (id, short_url, original_url, user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING id, short_url
	`

	var returnedShortURL string
	err := s.db.QueryRow(query, id, shortURL, originalURL, userID).Scan(&id, &returnedShortURL)

	if err != nil {
		existingShortURL, fetchErr := s.getShortURLByOriginal(originalURL)
		if fetchErr != nil {
			return "", fmt.Errorf("could not fetch existing short URL: %w", fetchErr)
		}
		return "", &ErrURLExists{ExistingShortURL: existingShortURL, ID: id}
	}

	return id, nil
}

func (s *PostgresStore) SaveBatch(items []models.BatchItem, userID *string) ([]URLData, error) {
	var urlDataList []URLData

	query := `
		INSERT INTO urls (id, short_url, original_url, user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING id, short_url
	`

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	for _, item := range items {
		var id, returnedShortURL string
		err := tx.QueryRow(query, item.CorrelationID, item.ShortURL, item.OriginalURL, userID).Scan(&id, &returnedShortURL)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				existingShortURL, fetchErr := s.getShortURLByOriginal(item.OriginalURL)
				if fetchErr != nil {
					return nil, fmt.Errorf("could not fetch conflicting URL: %w", fetchErr)
				}
				// Append existing URL as part of response.
				urlDataList = append(urlDataList, URLData{
					UUID:        item.CorrelationID,
					ShortURL:    existingShortURL,
					OriginalURL: item.OriginalURL,
				})
				continue
			}
			return nil, fmt.Errorf("could not insert batch item: %w", err)
		}

		urlDataList = append(urlDataList, URLData{
			UUID:        id,
			ShortURL:    returnedShortURL,
			OriginalURL: item.OriginalURL,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return urlDataList, nil
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

func (s *PostgresStore) GetUserURLs(userID string) ([]models.UserURLs, error) {
	var userURLs []models.UserURLs
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user URLs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, originalURL string
		if err := rows.Scan(&key, &originalURL); err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		shortURL, err := utils.ConstructURL(config.Config.BaseURL, key)
		if err != nil {
			return nil, fmt.Errorf("could not construct URL: %w", err)
		}
		userURLs = append(userURLs, models.UserURLs{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return userURLs, nil
}

func (s *PostgresStore) LoadFromFile(_ string) error {
	// не поддерживаем загрузку из файла
	return nil
}

func (s *PostgresStore) SaveToFile(_ string) error {
	// не поддерживаем сохранение в файл
	return nil
}

func (s *PostgresStore) getShortURLByOriginal(originalURL string) (string, error) {
	var shortURL string
	query := `SELECT short_url FROM urls WHERE original_url = $1`
	err := s.db.QueryRow(query, originalURL).Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("could not fetch short URL by original URL: %w", err)
	}
	return shortURL, nil
}
