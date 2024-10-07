package app

import (
	"fmt"
	"net/http"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/go-chi/chi/v5"
)

func Server() error {
	config.InitConfig()

	r := chi.NewRouter()

	r.Mount("/", ShortenerRouter())

	fmt.Printf("starting server at :%s\n", config.Addr)
	return http.ListenAndServe(config.Addr, r)
}
