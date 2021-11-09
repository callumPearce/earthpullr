package reddit_cli

import (
	"context"
	"earthpullr/internal/reddit_oauth"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
	"earthpullr/pkg/file_readers"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type BackgroundRetriever struct {
	logger                     *zap.Logger
	configMan                  config.ConfigManager
	secretsMan                 secrets.SecretsManager
	maxAggregatedQueryTimeSecs int
	runtime                    *wails.Runtime
	ctx                        context.Context
	client                     *http.Client
}

type BackgroundsRequest struct {
	Width          int
	Height         int
	BackgroundsCount int
	DownloadPath   string
}

func NewBackgroundRetriever(ctx context.Context, logger *zap.Logger, cm config.ConfigManager, sm secrets.SecretsManager) (*BackgroundRetriever, error) {
	backgroundConf, err := cm.GetMultiConfig([]string{
		"subreddit",
		"subreddit_search_type",
		"query_batch_size",
		"max_aggregated_query_time_secs",
	})
	if err != nil {
		return &BackgroundRetriever{}, err
	}
	maxAggregatedQueryTimeSecs, err := strconv.Atoi(backgroundConf["max_aggregated_query_time_secs"])
	if err != nil {
		err = fmt.Errorf("failed to parse config variable to integer: %w", err)
		return &BackgroundRetriever{}, err
	}
	retriever := &BackgroundRetriever{
		logger:                     logger,
		configMan:                  cm,
		secretsMan:                 sm,
		maxAggregatedQueryTimeSecs: maxAggregatedQueryTimeSecs,
		ctx:                        ctx,
		client:                     &http.Client{Timeout: 10 * time.Second},
	}
	return retriever, nil
}

func (br *BackgroundRetriever) WailsInit(runtime *wails.Runtime) error {
	br.runtime = runtime
	return nil
}

func (br *BackgroundRetriever) GetBackgrounds(request map[string]interface{}) (string, error) {
	var brRequest BackgroundsRequest
	err := mapstructure.Decode(request, &brRequest)
	if err != nil {
		return "Error", fmt.Errorf("failed to decode backgrounds request from frontend: %v", err)
	}
	if _, dirErr := os.Stat(brRequest.DownloadPath); os.IsNotExist(dirErr) {
		return "Error", fmt.Errorf("Download path '%s' does not exist", brRequest.DownloadPath)
	}
	br.logger.Info(fmt.Sprintf(
		"Received a request to retrieve %d backgrounds with a minimum resolution of %dx%d to directory %s",
		brRequest.BackgroundsCount,
		brRequest.Width,
		brRequest.Height,
		brRequest.DownloadPath,
	))
	err = br.addOAuthTokenToCtx()
	if err != nil {
		return "Error", fmt.Errorf("failed to get new backgrounds: %v", err)
	}
	existingBackgrounds := br.getExistingBackgroundsMap(brRequest.DownloadPath)
	err = br.getBackgroundsWithBatching(brRequest, existingBackgrounds)
	if err != nil {
		return "Error", err
	}
	return "Success", nil
}

func (br *BackgroundRetriever) getExistingBackgroundPath(downloadPath string) (string, error){
	existingImagesFname, err := br.configMan.GetConfig("existing_images_filename")
	if err != nil {
		return "", fmt.Errorf("failed to build path to existing images: %v", err)
	}
	existingBackgroundFpath := filepath.Join(downloadPath, existingImagesFname)
	return existingBackgroundFpath, nil
}

func (br *BackgroundRetriever) getExistingBackgroundsMap(downloadPath string) *map[string]string {
	existingBackgrounds := map[string]string{}
	existingBackgroundFpath, err := br.getExistingBackgroundPath(downloadPath)
	if err != nil {
		br.logger.Error(fmt.Sprintf("%v", err))
		return &existingBackgrounds
	}

	if _, err := os.Stat(existingBackgroundFpath); errors.Is(err, os.ErrNotExist) {
		// Existing backgrounds file doesn't exist, just return an empty map
		return &existingBackgrounds
	}

	existingBackgroundsJson, err := file_readers.NewFlatJsonFile(existingBackgroundFpath)
	if err != nil {
		br.logger.Error("failed to read existing images json file", zap.Error(err))
		return &existingBackgrounds
	}

	existingBackgrounds = existingBackgroundsJson.Data
	return &existingBackgrounds
}

func (br *BackgroundRetriever) saveExistingBackgrounds(downloadPath string, existingBackgrounds *map[string]string) error {
	existingBackgroundFpath, err := br.getExistingBackgroundPath(downloadPath)
	if err != nil {
		return err
	}
	err = file_readers.SaveMapAsJson(*existingBackgrounds, existingBackgroundFpath)
	br.logger.Info(fmt.Sprintf("saved existing backgrounds file to '%s'", existingBackgroundFpath))
	return err
}

func (br *BackgroundRetriever) getBackgroundsWithBatching(brRequest BackgroundsRequest, existingBackgrounds *map[string]string) error {
	savedImages := 0
	afterUID := ""
	for savedImages < brRequest.BackgroundsCount {
		listingRequest, err := NewListingRequest(
			br.ctx,
			br.client,
			br.configMan,
			"",
			afterUID,
		)
		if err != nil {
			return fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		listingResponse, err := listingRequest.DoRequest()
		if err != nil {
			return fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		remainingImagesCount := brRequest.BackgroundsCount - savedImages
		imagesRetriever, err := NewImagesRetriever(br.logger, br.ctx, listingResponse, br.client, remainingImagesCount, brRequest.Width, brRequest.Height, existingBackgrounds)
		afterUID = imagesRetriever.finalImageUID
		if err != nil {
			err = fmt.Errorf("failed to retrieve image batch: %v", err)
			return err
		}
		imagesRetriever.SaveImages(brRequest.DownloadPath, br.runtime, existingBackgrounds)
		savedImages += imagesRetriever.imageCount
	}
	return br.saveExistingBackgrounds(brRequest.DownloadPath, existingBackgrounds)
}

func (br *BackgroundRetriever) addOAuthTokenToCtx() error {
	redditOauth, err := reddit_oauth.NewApplicationOnlyOAuthRequest(br.ctx, br.client, br.secretsMan, br.configMan)
	if err != nil {
		return fmt.Errorf("failed to build request to retrieve oauth Token from reddit: %v", err)
	}
	oauthToken, err := redditOauth.NewOAuthToken()
	if err != nil {
		return fmt.Errorf("failed to retrieve oauth Token from reddit: %v", err)
	}
	br.ctx = reddit_oauth.ToContext(br.ctx, oauthToken)
	return nil
}
