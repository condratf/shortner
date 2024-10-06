package app

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getHandler(w, r)
	case http.MethodPost:
		postHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func Server() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler)

	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(`:8080`, mux)

	return err
}
