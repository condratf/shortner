package config

import (
	"flag"
	"os"
)

type config struct {
	Addr        string
	BaseURL     string
	FilePath    string
	DatabaseDSN string
}

var Config = config{
	Addr:        "localhost:8080",
	BaseURL:     "http://localhost:8080",
	FilePath:    "./shortener.json",
	DatabaseDSN: "",
}

func InitConfig() {
	addr := flag.String("a", "", "HTTP server address")
	baseURL := flag.String("b", "", "Base URL for shortened URL")
	filePath := flag.String("f", "", "Path to file for storing URLs in JSON format")
	databaseDSN := flag.String("d", "", "Database DSN")

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

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		Config.FilePath = envFilePath
	} else if *filePath != "" {
		Config.FilePath = *filePath
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		Config.DatabaseDSN = envDatabaseDSN
	} else if *databaseDSN != "" {
		Config.DatabaseDSN = *databaseDSN
	}
}
