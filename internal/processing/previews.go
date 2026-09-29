package processing

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

var UnsupportedMediaType = errors.New("Unsupported media type for preview generation")
var PreviewGenerationErr = errors.New("Failed to generate preview")
var FileWriteErr = errors.New("Failed to write preview file to disk")

type Video struct{}
type Image struct{}

func isVideo(path string) bool {

	cmd := exec.Command("ffprobe", "-v", "error", "-select-streams", "v:0", "-show-entries", "stream=codec_type", "-of", "csv=p=0", path)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "video"
}

func getMediaType(imagePath string) (interface{}, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buffer := make([]byte, 128)
	_, err = file.Read(buffer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	imageType := vips.DetermineImageType(buffer)
	if imageType == vips.ImageTypeUnknown {
		if !isVideo(imagePath) {
			return nil, UnsupportedMediaType
		} else {
			return Video{}, nil
		}
	}
	return Image{}, nil
}

func generateImagePreview(imagePath string) (string, error) {
	image, err := vips.NewImageFromFile(imagePath)
	if err != nil {
		log.Println(err)
		return "", err
	}

	webpParams := vips.NewWebpExportParams()
	webpParams.Quality = 25
	webpParams.StripMetadata = true
	webpParams.Lossless = false

	buffer, _, err := image.ExportWebp(webpParams)
	if err != nil {
		log.Println(PreviewGenerationErr)
		return "", PreviewGenerationErr
	}
	fileName := filepath.Base(imagePath)

	newFile := filepath.Join(filepath.Dir(imagePath), strings.TrimSuffix(fileName, filepath.Ext(fileName))+"_preview.webp")

	err = os.WriteFile(newFile, buffer, 0644)
	if err != nil {
		log.Println(FileWriteErr)
		return "", FileWriteErr

	}
	return newFile, nil
}

func generateVideoPreview(videoPath string) (string, error) {

	fileName := filepath.Base(videoPath)
	newFile := filepath.Join(filepath.Dir(videoPath), strings.TrimSuffix(fileName, filepath.Ext(fileName))+"_preview.webp")

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", newFile)

	_, err := cmd.Output()
	if err != nil {
		log.Println(PreviewGenerationErr)
		return "", PreviewGenerationErr
	}
	return newFile, nil
}

func GeneratePreview(filePath string) error {

	mediaType, err := getMediaType(filePath)
	if err != nil {
		log.Println(err)
		return err
	}

	switch mediaType {
	case Image{}:
		path, err := generateImagePreview(filePath)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Generated image preview at ", path)
		return nil

	case Video{}:

		path, err := generateVideoPreview(filePath)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Generated video preview at ", path)
		return nil
	}
	return nil
}
