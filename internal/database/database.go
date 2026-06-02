package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type DataStore struct {
	db *pgxpool.Pool
}

func New() *DataStore {
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
	return &DataStore{db: Pool}
}

func (ds *DataStore) AddUser(username, password, email string) error {
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = ds.db.Exec(context.Background(),
		"INSERT INTO users (password_hash, email) VALUES ($1, $2)",
		string(bcryptPassword),
		email)
	return err
}

func (ds *DataStore) getUser(email string) (int, string, error) {
	var passwordHash string
	var id int
	err := ds.db.QueryRow(context.Background(),
		"SELECT password_hash, id FROM users WHERE email = $1",
		email).Scan(&passwordHash, &id)
	return id, passwordHash, err
}
func (ds *DataStore) Login(email, password string) (int, error) {
	userID, passwordHash, err := ds.getUser(email)
	if err != nil {
		fmt.Println("Error getting user:", err)
		if err == pgx.ErrNoRows {
			return -1, AccountNotFoundError
		}
		return -1, err

	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return -1, InvalidPasswordError
	}
	return userID, nil
}
func (ds *DataStore) Close() {
	ds.db.Close()
}
