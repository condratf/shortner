package errorhandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/condratf/shortner/internal/app/utils"
)

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
