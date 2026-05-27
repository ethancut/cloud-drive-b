package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New() *pgxpool.Pool {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set!")
	}
	var err error
	var Pool *pgxpool.Pool
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	return Pool
}
