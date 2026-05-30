package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

func AddUser(Pool *pgxpool.Pool, username, password, email string) error {
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = Pool.Exec(context.Background(),
		"INSERT INTO users (password_hash, email) VALUES ($1, $2)",
		string(bcryptPassword),
		email)
	return err
}

func getUser(Pool *pgxpool.Pool, email string) (string, error) {
	var passwordHash string
	err := Pool.QueryRow(context.Background(),
		"SELECT password_hash FROM users WHERE email = $1",
		email).Scan(&passwordHash)
	return passwordHash, err
}
func Login(Pool *pgxpool.Pool, email, password string) (bool, error) {
	passwordHash, err := getUser(Pool, email)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}
