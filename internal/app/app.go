package app

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Server() error {
	r := chi.NewRouter()

	r.Mount("/", ShortenerRouter())

	fmt.Println("starting server at :8080")
	return http.ListenAndServe(":8080", r)
}
