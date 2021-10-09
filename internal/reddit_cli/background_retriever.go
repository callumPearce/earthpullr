package reddit_cli

import (
	"context"
	"earthpullr/internal/reddit_oauth"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"net/http"
	"os"
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
		return "Error", fmt.Errorf("download path '%s' does not exist", brRequest.DownloadPath)
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
	err = br.getBackgroundsWithBatching(brRequest)
	if err != nil {
		return "Error", err
	}
	return "Success", nil
}

func (br *BackgroundRetriever) getBackgroundsWithBatching(brRequest BackgroundsRequest) error {
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
		imagesRetriever, err := NewImagesRetriever(br.logger, br.ctx, listingResponse, br.client, remainingImagesCount, brRequest.Width, brRequest.Height)
		afterUID = imagesRetriever.finalImageUID
		if err != nil {
			err = fmt.Errorf("failed to retrieve image batch: %v", err)
			return err
		}
		imagesRetriever.SaveImages(brRequest.DownloadPath, br.runtime)
		savedImages += imagesRetriever.imageCount
	}
	return nil
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
