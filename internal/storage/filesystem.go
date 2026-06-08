package storage

import (
	"fmt"
	"io"
	"os"
	"time"
)

type File struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modtime"`
}

func UploadFile(userID int, filename string, file io.Reader) error {
	dir := fmt.Sprintf("%s%d", os.Getenv("UPLOADS_DIR"), userID)
	os.MkdirAll(dir, os.ModePerm)
	dst, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	return err
}

func ListFiles(userID int) ([]File, error) {
	dir := fmt.Sprintf("%s%d", os.Getenv("UPLOADS_DIR"), userID)
	files := []File{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, File{
				Filename: entry.Name(),
				Size:     info.Size(),
				ModTime:  info.ModTime(),
			})
		}
	}
	return files, nil
}
