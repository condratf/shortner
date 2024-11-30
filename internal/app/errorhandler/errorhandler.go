package errorhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/storage"
	"github.com/condratf/shortner/internal/app/utils"
)

const (
	ResponseTypeJSON      = "json"
	ResponseTypeJSONBatch = "json-batch"
	ResponseTypeText      = "text"
)

type responsePayload struct {
	Result string `json:"result"`
}

func HandleURLExistError(w http.ResponseWriter, err error, respType string) bool {
	if errors.Is(err, &storage.ErrURLExists{}) {
		var urlExistsErr *storage.ErrURLExists
		if errors.As(err, &urlExistsErr) {
			shortURL, err := constructShortURL(urlExistsErr.ExistingShortURL, w)
			if err != nil {
				return true
			}
			switch respType {
			case ResponseTypeJSON:
				writeJSONResponse(w, responsePayload{Result: shortURL})
			case ResponseTypeJSONBatch:
				writeJSONResponse(w, models.ResponsePayloadBatch{
					CorrelationID: urlExistsErr.ID,
					ShortURL:      shortURL,
				})
			case ResponseTypeText:
				writeTextResponse(w, shortURL)
			}
			return true
		}
	}
	return false
}

func constructShortURL(existingShortURL string, w http.ResponseWriter) (string, error) {
	shortURL, err := utils.ConstructURL(config.Config.BaseURL, existingShortURL)
	if err != nil {
		http.Error(w, "could not construct URL", http.StatusInternalServerError)
		return "", err
	}
	return strings.TrimSpace(shortURL), nil
}

func writeJSONResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeTextResponse(w http.ResponseWriter, shortURL string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusConflict)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		http.Error(w, "could not write response", http.StatusInternalServerError)
	}
}
