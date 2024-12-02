package models

type RequestPayloadBatch struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ResponsePayloadBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BatchItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
}

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortURLAndStore = func(originalURL string, userID *string) (string, error)

type ShortURLAndStoreBatch = func(items []RequestPayloadBatch, userID *string) ([]BatchItem, error)
