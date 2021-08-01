package reddit_cli

import (
	"context"
	"earthpullr/src/reddit_oauth"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MAX_RES = 7680 // 8K
const ACCEPTABLE_ASPECT_DIFF = 0.25

type ListingsImagesRetriever struct {
	requests      map[imageData]*http.Request
	oAuthToken    *reddit_oauth.OAuthToken
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

func (retriever ListingsImagesRetriever) SaveImages(directoryPath string) (err error) {
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
		log.Info(fmt.Sprintf("Successfully saved image to '%s'", filePath))
	}
	return nil
}

func imageAboveMinSize(image imageData, width int, height int) (valid bool) {
	valid = true
	if image.Width < width || image.Height < height {
		valid = false
		log.Info(fmt.Sprintf(
			"Image with found with resolution (%d, %d) does not meet minimum (%d, %d)",
			image.Width,
			image.Height,
			width,
			height,
		))
	}
	return valid
}

func imageWithinAspectRatioRange(image imageData, width int, height int) (valid bool) {
	valid = true
	aspectRatio := (float64(image.Width) / float64(image.Height))
	requiredRatio := (float64(width) / float64(height))
	aspectDiff := aspectRatio - requiredRatio
	if aspectDiff < 0.0 {
		aspectDiff = -aspectDiff
	}
	if aspectDiff > float64(ACCEPTABLE_ASPECT_DIFF) {
		valid = false
		log.Info(fmt.Sprintf(
			"Image found with aspect ratio %f, required (+/-%f)%f",
			aspectRatio,
			float64(ACCEPTABLE_ASPECT_DIFF),
			requiredRatio,
		))
	}
	return valid
}

func imageFitsSpecifiedResolution(image imageData, width int, height int) (valid bool) {
	valid = imageAboveMinSize(image, width, height)
	valid = valid && imageWithinAspectRatioRange(image, width, height)
	return valid
}

func NewImagesRetriever(lres ListingResponse, oAuthToken *reddit_oauth.OAuthToken, client *http.Client, width int, height int) (imagesRetriever ListingsImagesRetriever, err error) {
	var images []imageData

	if width <= 0 || width > MAX_RES || height <= 0 || height > MAX_RES {
		return imagesRetriever, fmt.Errorf("resolution must be between (1, 1) to (%d, %d), got (%d, %d)", MAX_RES, MAX_RES, width, height)
	}

	for _, child := range lres.Data.Children {
		image := imageData{
			UID:   child.Data.Name,
			Title: child.Data.Title,
		}
		for _, imageObj := range child.Data.Preview.ImagesList {
			image.URL = imageObj.Source.URL
			image.Width = imageObj.Source.Width
			image.Height = imageObj.Source.Height
			if imageFitsSpecifiedResolution(image, width, height) {
				images = append(images, image)
			}
		}
	}
	requests := map[imageData]*http.Request{}
	for i, image := range images {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			html.UnescapeString(image.URL),
			nil,
		)
		if err != nil {
			err = fmt.Errorf("failed to create request to retrieve image for url '%s', reason: %v", image.URL, err)
			return imagesRetriever, err
		}
		requests[image] = req
		if i == len(images)-1 {
			imagesRetriever.finalImageUID = image.UID
		}
	}
	imagesRetriever.requests = requests
	imagesRetriever.oAuthToken = oAuthToken
	imagesRetriever.client = client
	imagesRetriever.imageCount = len(requests)
	return imagesRetriever, err
}
