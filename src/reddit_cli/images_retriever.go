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

type ImagesRetriever struct {
	requests   map[imageData]*http.Request
	oAuthToken *reddit_oauth.OAuthToken
	client     *http.Client
}

type imageData struct {
	URL   string
	Title string
	UID   string
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

func (retriever ImagesRetriever) saveResponseToFile(filePath string, res *http.Response) error {
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

func (retriever ImagesRetriever) SaveImages(directoryPath string) (err error) {
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

func NewImagesRetriever(lres ListingResponse, oAuthToken *reddit_oauth.OAuthToken, client *http.Client) (imagesRetriever ImagesRetriever, err error) {
	var images []imageData
	for _, child := range lres.Data.Children {
		image := imageData{
			UID:   child.Data.Name,
			Title: child.Data.Title,
		}
		for _, imageObj := range child.Data.Preview.ImagesList {
			image.URL = imageObj.Source.URL
			images = append(images, image)
		}
	}
	requests := map[imageData]*http.Request{}
	for _, image := range images {
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
	}
	imagesRetriever.requests = requests
	imagesRetriever.oAuthToken = oAuthToken
	imagesRetriever.client = client
	return imagesRetriever, err
}
