package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/condratf/shortner/internal/app/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", config.Config.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("could not connect to the database: %w", err)
	}

	DB.SetConnMaxLifetime(time.Minute * 3)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("could not ping the database: %w", err)
	}

	return nil
}

func PingDB(ctx context.Context) error {
	return DB.PingContext(ctx)
}

func CloseDB() error {
	return DB.Close()
}
