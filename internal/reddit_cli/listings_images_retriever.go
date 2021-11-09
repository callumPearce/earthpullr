package reddit_cli

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const MAX_RES = 7680 // 8K
const ACCEPTABLE_ASPECT_DIFF = 0.25

type ListingsImagesRetriever struct {
	logger        *zap.Logger
	requests      map[imageData]*http.Request
	client        *http.Client
	imageCount    int
	finalImageUID string
}

type imageData struct {
	URL    string
	Title  string
	UID    string
	Width  int
	Height int
}

func (image imageData) getImageFileType() (string, error) {
	switch url := image.URL; {
	case strings.Contains(url, ".jpg"):
		return ".jpg", nil
	case strings.Contains(url, ".png"):
		return ".png", nil
	default:
		return "", fmt.Errorf("unknown file type for image with url: %s", url)
	}
}

func (image imageData) getImageName() (fileType string, err error) {
	fileType, err = image.getImageFileType()
	return image.UID + fileType, err
}

func (retriever ListingsImagesRetriever) saveResponseToFile(filePath string, res *http.Response) error {
	defer res.Body.Close()
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s', reason: %v", filePath, err)
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return fmt.Errorf("failed to save bytes to file '%s', reason: %v", filePath, err)
	}
	return nil
}

func (retriever ListingsImagesRetriever) SaveImages(directoryPath string, runtime *wails.Runtime, existingBackgrounds *map[string]string) error {
	for image, request := range retriever.requests {
		fileName, err := image.getImageName()
		if err != nil {
			return fmt.Errorf("failed to save image locally for url '%s': %v", image.URL, err)
		}
		filePath := filepath.Join(directoryPath, fileName)

		res, err := retriever.client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to download with URL '%s', reason: %v", image.URL, err)
		}
		err = retriever.saveResponseToFile(filePath, res)
		if err != nil {
			return err
		}
		retriever.logger.Info(fmt.Sprintf("Successfully saved image to '%s'", filePath))
		(*existingBackgrounds)[fileName] = "s"
		runtime.Events.Emit("image_saved", 1)
	}
	return nil
}

func imageAboveMinSize(logger *zap.Logger, image imageData, width int, height int) (valid bool) {
	valid = true
	if image.Width < width || image.Height < height {
		valid = false
		logger.Debug(fmt.Sprintf(
			"Image with found with resolution (%d, %d) does not meet minimum (%d, %d)",
			image.Width,
			image.Height,
			width,
			height,
		))
	}
	return valid
}

func imageWithinAspectRatioRange(logger *zap.Logger, image imageData, width int, height int) (valid bool) {
	valid = true
	aspectRatio := (float64(image.Width) / float64(image.Height))
	requiredRatio := (float64(width) / float64(height))
	aspectDiff := aspectRatio - requiredRatio
	if aspectDiff < 0.0 {
		aspectDiff = -aspectDiff
	}
	if aspectDiff > float64(ACCEPTABLE_ASPECT_DIFF) {
		valid = false
		logger.Debug(fmt.Sprintf(
			"Image found with aspect ratio %f, required (+/-%f)%f",
			aspectRatio,
			float64(ACCEPTABLE_ASPECT_DIFF),
			requiredRatio,
		))
	}
	return valid
}

func imageFitsSpecifiedResolution(logger *zap.Logger, image imageData, width int, height int) (valid bool) {
	valid = imageAboveMinSize(logger, image, width, height)
	valid = valid && imageWithinAspectRatioRange(logger, image, width, height)
	return valid
}

func imageHasBeenDownloaded(logger *zap.Logger, image imageData, existingImages *map[string]string) (exists bool) {
	fname, err := image.getImageName()
	if err != nil {
		return false
	}
	if _, ok := (*existingImages)[fname]; ok {
		logger.Debug(fmt.Sprintf("Image '%s' already exists in the download directory", fname, ))
		return true
	}
	return false
}

func NewImagesRetriever(logger *zap.Logger, ctx context.Context, lres ListingResponse, client *http.Client, maxImages int, width int, height int, existingImages *map[string]string) (imagesRetriever ListingsImagesRetriever, err error) {
	var images []imageData

	if width <= 0 || width > MAX_RES || height <= 0 || height > MAX_RES {
		return imagesRetriever, fmt.Errorf("resolution must be between (1, 1) to (%d, %d), got (%d, %d)", MAX_RES, MAX_RES, width, height)
	}

	imageCount := 0
	for _, child := range lres.Data.Children {
		if imageCount >= maxImages {
			break
		}
		image := imageData{
			UID:   child.Data.Name,
			Title: child.Data.Title,
		}
		imagesRetriever.finalImageUID = image.UID
		for _, imageObj := range child.Data.Preview.ImagesList {
			image.URL = imageObj.Source.URL
			image.Width = imageObj.Source.Width
			image.Height = imageObj.Source.Height
			if imageFitsSpecifiedResolution(logger, image, width, height) && !imageHasBeenDownloaded(logger, image, existingImages) {
				images = append(images, image)
				imageCount += 1
			}
		}
	}
	requests := map[imageData]*http.Request{}
	for _, image := range images {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			html.UnescapeString(image.URL),
			nil,
		)
		if err != nil {
			err = fmt.Errorf("failed to create request to retrieve image for url '%s', reason: %v", image.URL, err)
			return imagesRetriever, err
		}
		requests[image] = req
	}
	imagesRetriever.logger = logger
	imagesRetriever.requests = requests
	imagesRetriever.client = client
	imagesRetriever.imageCount = len(requests)
	return imagesRetriever, err
}
