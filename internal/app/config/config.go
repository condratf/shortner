package config

import (
	"flag"
	"os"
)

type config struct {
	Addr    string
	BaseURL string
}

var Config = config{
	Addr:    "localhost:8080",
	BaseURL: "http://localhost:8080",
}

func InitConfig() {
	addr := flag.String("a", "", "HTTP server address")
	baseURL := flag.String("b", "", "Base URL for shortened URL")

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		Config.Addr = envAddr
	} else if *addr != "" {
		Config.Addr = *addr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		Config.BaseURL = envBaseURL
	} else if *baseURL != "" {
		Config.BaseURL = *baseURL
	}
}
