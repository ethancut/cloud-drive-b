package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
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

	id := uuid.New()

	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = ds.db.Exec(context.Background(),
		"INSERT INTO users (id, password_hash, email, username) VALUES ($1, $2, $3, $4)",
		id,
		string(bcryptPassword),
		email, username)
	return err
}

func (ds *DataStore) getUser(email string) (string, string, string, error) {
	var passwordHash, username string
	var id string
	err := ds.db.QueryRow(context.Background(),
		"SELECT password_hash, id, username FROM users WHERE email = $1",
		email).Scan(&passwordHash, &id, &username)
	return id, passwordHash, username, err
}
func (ds *DataStore) DeleteUser(email string) error {
	_, err := ds.db.Exec(context.Background(),
		"DELETE FROM users WHERE email = $1",
		email)
	return err
}

func (ds *DataStore) Login(email, password string) (uuid.UUID, string, error) {
	userID, passwordHash, username, err := ds.getUser(email)
	if err != nil {
		fmt.Println("Error getting user:", err)
		if err == pgx.ErrNoRows {
			return uuid.Nil, "", AccountNotFoundError
		}
		return uuid.Nil, "", err

	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return uuid.Nil, "", InvalidPasswordError
	}
	return uuid.MustParse(userID), username, nil
}
func (ds *DataStore) Close() {
	ds.db.Close()
}
func (ds *DataStore) AddFile(ctx context.Context, id uuid.UUID, filename string, path string, userID uuid.UUID, size int64) error {
	_, err := ds.db.Exec(ctx,
		"INSERT INTO files (id, original_filename, filepath, user_id, size_bytes) VALUES ($1, $2, $3, $4, $5)",
		id, filename, path, userID, size)
	return err
}

type FileMetadata struct {
	ID               uuid.UUID
	OriginalFilename string
	FilePath         string
	UserID           uuid.UUID
	SizeBytes        int64
	CreatedAt        time.Time
}

func (ds *DataStore) GetAllFileMetadata(ctx context.Context, userID uuid.UUID) ([]FileMetadata, error) {

	rows, err := ds.db.Query(ctx,
		"SELECT id, original_filename, filepath, size_bytes, created_at FROM files WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []FileMetadata{}
	for rows.Next() {
		var f FileMetadata
		if err := rows.Scan(&f.ID, &f.OriginalFilename, &f.FilePath, &f.SizeBytes, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
func (ds *DataStore) GetFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*FileMetadata, error) {
	var f FileMetadata
	err := ds.db.QueryRow(ctx,
		"SELECT id, original_filename, filepath, size_bytes, created_at FROM files WHERE id = $1 AND user_id = $2",
		fileID, userID).Scan(&f.ID, &f.OriginalFilename, &f.FilePath, &f.SizeBytes, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	log.Println(f)
	return &f, nil
}

func (ds *DataStore) DeleteFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (string, error) {
	var path string
	err := ds.db.QueryRow(ctx,
		"DELETE FROM files WHERE id = $1 AND user_id = $2 RETURNING filepath",
		fileID, userID).Scan(&path)
	return path, err
}
