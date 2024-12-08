package errorhandler

import (
	"errors"
	"net/http"

	"github.com/condratf/shortner/internal/app/models"
	"github.com/condratf/shortner/internal/app/storage"
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
	if !errors.Is(err, &storage.ErrURLExists{}) {
		return false
	}
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
	return false
}
