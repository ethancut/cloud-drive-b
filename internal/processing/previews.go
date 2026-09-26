package processing

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

var FileNotFoundErr = errors.New("Unsupported or invalid file type for preview generation")
var PreviewGenerationErr = errors.New("Failed to generate preview")
var FileWriteErr = errors.New("Failed to write preview file to disk")

func GeneratePreview(imagePath string) {
	image, err := vips.NewImageFromFile(imagePath)
	if err != nil {
		log.Println("")
		return
	}
	fileName := filepath.Base(imagePath)
	webpParams := vips.NewWebpExportParams()
	webpParams.Quality = 25
	webpParams.StripMetadata = true
	webpParams.Lossless = false

	buffer, _, err := image.ExportWebp(webpParams)
	if err != nil {
		log.Println(PreviewGenerationErr)
		return
	}

	newFile := filepath.Join(filepath.Dir(imagePath), strings.TrimSuffix(fileName, filepath.Ext(fileName))+"_preview.webp")

	err = os.WriteFile(newFile, buffer, 0644)
	if err != nil {
		log.Println(FileWriteErr)
		return
	}
}
