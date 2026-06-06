package storage

import (
	"fmt"
	"io"
	"os"
)

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
