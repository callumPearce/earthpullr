package reddit_cli

import (
	"context"
	"earthpullr/internal/reddit_oauth"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
	"fmt"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type BackgroundRetriever struct {
	logger     *zap.Logger
	configMan  config.ConfigManager
	secretsMan secrets.SecretsManager
	width      int
	height                     int
	maxAggregatedQueryTimeSecs int
	backgroundsCount           int
	runtime *wails.Runtime
	ctx context.Context
}

func addOAuthTokenToCtx(ctx context.Context, client *http.Client, sm secrets.SecretsManager, cm config.ConfigManager) (context.Context, error) {
	redditOauth, err := reddit_oauth.NewApplicationOnlyOAuthRequest(ctx, client, sm, cm)
	if err != nil {
		return ctx, fmt.Errorf("failed to build request to retrieve oauth Token from reddit: %v", err)
	}
	oauthToken, err := redditOauth.NewOAuthToken()
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve oauth Token from reddit: %v", err)
	}
	ctx = reddit_oauth.ToContext(ctx, oauthToken)
	return ctx, nil
}

func (br *BackgroundRetriever) WailsInit(runtime *wails.Runtime) error {
	br.runtime = runtime
	return nil
}

	func (br *BackgroundRetriever) GetBackgrounds() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	ctx, err := addOAuthTokenToCtx(br.ctx, client, br.secretsMan, br.configMan)
	if err != nil {
		fmt.Errorf("failed to get new backgrounds: %v", err)
	}
	savedImages := 0
	afterUID := ""

	for savedImages < br.backgroundsCount {
		listingRequest, err := NewListingRequest(
			ctx,
			client,
			br.configMan,
			"",
			afterUID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		listingResponse, err := listingRequest.DoRequest()
		if err != nil {
			return "", fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		remainingImagesCount := br.backgroundsCount - savedImages
		imagesRetriever, err := NewImagesRetriever(br.logger, ctx, listingResponse, client, remainingImagesCount, br.width, br.height)
		afterUID = imagesRetriever.finalImageUID
		if err != nil {
			err = fmt.Errorf("failed to retrieve image batch: %v", err)
			return "", err
		}
		imagesRetriever.SaveImages("images", br.runtime)
		savedImages += imagesRetriever.imageCount
	}
	return "Success", nil
}

func NewBackgroundRetriever(ctx context.Context, logger *zap.Logger, cm config.ConfigManager, sm secrets.SecretsManager) (*BackgroundRetriever, error) {
	backgroundConf, err := cm.GetMultiConfig([]string{
		"subreddit",
		"subreddit_search_type",
		"width",
		"height",
		"query_batch_size",
		"max_aggregated_query_time_secs",
		"backgrounds_count",
	})
	if err != nil {
		return &BackgroundRetriever{}, err
	}
	width, err := strconv.Atoi(backgroundConf["width"])
	if err != nil {
		err = fmt.Errorf("failed to parse config variable to integer: %w", err)
		return &BackgroundRetriever{}, err
	}
	height, err := strconv.Atoi(backgroundConf["height"])
	if err != nil {
		err = fmt.Errorf("failed to parse config variable to integer: %w", err)
		return &BackgroundRetriever{}, err
	}
	maxAggregatedQueryTimeSecs, err := strconv.Atoi(backgroundConf["max_aggregated_query_time_secs"])
	if err != nil {
		err = fmt.Errorf("failed to parse config variable to integer: %w", err)
		return &BackgroundRetriever{}, err
	}
	backgroundsCount, err := strconv.Atoi(backgroundConf["backgrounds_count"])
	if err != nil {
		err = fmt.Errorf("failed to parse config variable to integer: %w", err)
		return &BackgroundRetriever{}, err
	}
	retriever := &BackgroundRetriever{
		logger: 					logger,
		configMan:                  cm,
		secretsMan:                 sm,
		width:                      width,
		height:                     height,
		maxAggregatedQueryTimeSecs: maxAggregatedQueryTimeSecs,
		backgroundsCount:           backgroundsCount,
		ctx:					 	ctx,
	}
	return retriever, nil
}
