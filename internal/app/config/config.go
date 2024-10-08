package config

import (
	"flag"
	"os"
)

var (
	Addr    string = "localhost:8080"
	BaseURL string = "http://localhost:8080"
)

func InitConfig() {
	addr := flag.String("a", "", "HTTP server address")
	baseURL := flag.String("b", "", "Base URL for shortened URL")

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		Addr = envAddr
	} else if *addr != "" {
		Addr = *addr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		BaseURL = envBaseURL
	} else if *baseURL != "" {
		BaseURL = *baseURL
	}
}
