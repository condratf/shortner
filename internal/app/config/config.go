package config

import (
	"flag"
)

var (
	Addr    string = "localhost:8080"
	BaseURL string = "http://localhost:8080"
)

func InitConfig() {
	addr := flag.String("a", Addr, "HTTP server address")
	baseURL := flag.String("b", BaseURL, "Base URL for shortened URL")

	flag.Parse()

	Addr = *addr
	BaseURL = *baseURL
}
