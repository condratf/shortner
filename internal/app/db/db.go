package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/condratf/shortner/internal/app/config"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

const (
	connectionMaxLifeTime = time.Minute * 3
	maxConnections        = 10
	defaultTimeout        = 5 * time.Second
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", config.Config.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("could not connect to the database: %w", err)
	}

	DB.SetConnMaxLifetime(connectionMaxLifeTime)
	DB.SetMaxOpenConns(maxConnections)
	DB.SetMaxIdleConns(maxConnections)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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
	log.Println("closing the database connection")
	return DB.Close()
}

func ApplyMigrations(dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	return nil
}
