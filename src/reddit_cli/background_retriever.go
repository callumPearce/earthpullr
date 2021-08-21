package reddit_cli

import (
	"context"
	"earthpullr/src/config"
	"earthpullr/src/reddit_oauth"
	"earthpullr/src/secrets"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type BackgroundRetriever struct {
	configMan                  config.ConfigManager
	secretsMan                 secrets.SecretsManager
	width                      int
	height                     int
	maxAggregatedQueryTimeSecs int
	backgroundsCount           int
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

func (br *BackgroundRetriever) GetBackgrounds(ctx context.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}
	ctx, err := addOAuthTokenToCtx(ctx, client, br.secretsMan, br.configMan)
	log.Info(reddit_oauth.FromContext(ctx))
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
			return fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		listingResponse, err := listingRequest.DoRequest()
		if err != nil {
			return fmt.Errorf("failed to get Listings for subreddit: %v", err)
		}
		remainingImagesCount := br.backgroundsCount - savedImages
		imagesRetriever, err := NewImagesRetriever(ctx, listingResponse, client, remainingImagesCount, br.width, br.height)
		afterUID = imagesRetriever.finalImageUID
		log.Info(afterUID)
		if err != nil {
			err = fmt.Errorf("failed retriever image batch: %v", err)
			return err
		}
		imagesRetriever.SaveImages("images")
		savedImages += imagesRetriever.imageCount
	}
	return nil
}

func NewBackgroundRetriever(cm config.ConfigManager, sm secrets.SecretsManager) (*BackgroundRetriever, error) {
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
		configMan:                  cm,
		secretsMan:                 sm,
		width:                      width,
		height:                     height,
		maxAggregatedQueryTimeSecs: maxAggregatedQueryTimeSecs,
		backgroundsCount:           backgroundsCount,
	}
	return retriever, nil
}
