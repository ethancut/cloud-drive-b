package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

<<<<<<< HEAD
	"github.com/ethannself/cloud-drive-b/internal/processing"
=======
	"github.com/google/uuid"
>>>>>>> origin/main
)

type File struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modtime"`
}

<<<<<<< HEAD
func UploadFile(userID int, filename string, file io.Reader) error {
	dir := fmt.Sprintf("%s%d", os.Getenv("UPLOADS_DIR"), userID)
	os.MkdirAll(dir, os.ModePerm)
	path := filepath.Join(dir, filename)

	dst, err := os.Create(path)
=======
func UploadFile(ctx context.Context, userID uuid.UUID, originalFilename string, file io.Reader) (uuid.UUID, error) {
	id := uuid.New()

	ext := filepath.Ext(originalFilename)
	storageFilename := id.String() + ext

	dir := os.Getenv("UPLOADS_DIR")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return uuid.UUID{}, err
	}

	storagePath := filepath.Join(dir, storageFilename)

	dst, err := os.Create(storagePath)
>>>>>>> origin/main
	if err != nil {
		return uuid.Nil, err
	}
	defer dst.Close()
<<<<<<< HEAD
	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	go processing.GeneratePreview(path)
	return err
=======

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(storagePath)
		return uuid.Nil, err
	}

	ds := GetDataStore()

	if err := ds.AddFile(ctx, id, originalFilename, storagePath, userID, written); err != nil {
		os.Remove(storagePath)
		return uuid.Nil, err
	}
	return id, nil
>>>>>>> origin/main
}

func ListFiles(ctx context.Context, userID uuid.UUID) ([]File, error) {
	ds := GetDataStore()
	filesMetadata, err := ds.GetAllFileMetadata(ctx, userID)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0, len(filesMetadata))
	for _, entry := range filesMetadata {
		files = append(files, File{
			ID:       entry.ID,
			Filename: entry.OriginalFilename,
			Size:     entry.SizeBytes,
			ModTime:  entry.CreatedAt,
		})
	}
	return files, nil
}

// if successful, returns the absolute path of the file.
func queryStorage(userID uuid.UUID, fileName string) (string, error) {
	dir := fmt.Sprintf("%s%s", os.Getenv("UPLOADS_DIR"), userID)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {

			if entry.Name() == fileName {
				absPath := filepath.Join(dir, entry.Name())
				return absPath, nil
			}

		}
	}
	return "", fmt.Errorf("file %q not found", fileName)
}
func DeleteFile(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) error {
	ds := GetDataStore()

	path, err := ds.DeleteFileMetadata(ctx, fileID, userID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
