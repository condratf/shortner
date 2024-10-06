package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
		http.Error(w, "could not read request body", http.StatusInternalServerError)
		return
	}

	url := string(reqBody)

	if len(url) > 0 {
		k, err := shortURLAndStore(url)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(k))
		return
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path

	parts := strings.Split(path, "/")
	if len(parts) > 1 && parts[1] != "" {
		id := parts[1]

		url, err := getURL(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
