package storage

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStore_SaveBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := &PostgresStore{db: db}

	items := []BatchItem{
		{CorrelationID: uuid.New().String(), ShortURL: "short1", OriginalURL: "http://example.com/1"},
		{CorrelationID: uuid.New().String(), ShortURL: "short2", OriginalURL: "http://example.com/2"},
	}

	query := `
			INSERT INTO urls (id, short_url, original_url)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING id, short_url
	`

	mock.ExpectBegin()

	for _, item := range items {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(item.CorrelationID, item.ShortURL, item.OriginalURL).
			WillReturnRows(sqlmock.NewRows([]string{"id", "short_url"}).AddRow(item.CorrelationID, item.ShortURL))
	}

	mock.ExpectCommit()

	urlDataList, err := store.SaveBatch(items)
	assert.NoError(t, err, "Expected no error during SaveBatch")
	assert.Len(t, urlDataList, len(items), "Expected urlDataList to have the same length as input items")

	for i, item := range items {
		assert.Equal(t, item.CorrelationID, urlDataList[i].UUID)
		assert.Equal(t, item.ShortURL, urlDataList[i].ShortURL)
		assert.Equal(t, item.OriginalURL, urlDataList[i].OriginalURL)
	}

	// Case: Transaction fails to begin
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
	_, err = store.SaveBatch(items)
	assert.Error(t, err, "Expected error when transaction fails to begin")

	// Case: Query fails
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(items[0].CorrelationID, items[0].ShortURL, items[0].OriginalURL).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.SaveBatch(items)
	assert.Error(t, err, "Expected error when query fails")
}
