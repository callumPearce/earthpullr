package reddit_cli

import (
	"earthpullr/src/config"
	"earthpullr/src/reddit_oauth"
	"earthpullr/src/secrets"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type BackgroundRetriever struct {
	configMan                  config.ConfigManager
	secretsMan                 secrets.SecretsManager
	width                      int
	height                     int
	maxAggregatedQueryTimeSecs int
	backgroundsCount           int
}

func getOAuthToken(client *http.Client, sm secrets.SecretsManager, cm config.ConfigManager) *reddit_oauth.OAuthToken {
	redditOauth, err := reddit_oauth.NewApplicationOnlyOAuthRequest(client, sm, cm)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to retrieve oauth Token from reddit: %v", err))
		os.Exit(1)
	}
	oauthToken, err := redditOauth.NewOAuthToken()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to retrieve oauth Token from reddit: %v", err))
		os.Exit(1)
	}
	return oauthToken
}

func (br *BackgroundRetriever) GetBackgrounds() error {
	client := &http.Client{Timeout: 10 * time.Second}
	oauthToken := getOAuthToken(client, br.secretsMan, br.configMan)

	savedImages := 0
	afterUID := ""

	for savedImages < br.backgroundsCount {
		listingRequest, err := NewListingRequest(
			client,
			oauthToken,
			br.configMan,
			"",
			afterUID,
		)
		if err != nil {
			err = fmt.Errorf("failed to get Listings for EarthPorn subreddit: %v", err)
			return err
		}
		listingResponse, err := listingRequest.DoRequest()
		if err != nil {
			err = fmt.Errorf("failed to get Listings for EarthPorn subreddit: %v", err)
			return err
		}
		imagesRetriever, err := NewImagesRetriever(listingResponse, oauthToken, client, br.width, br.height)
		afterUID = imagesRetriever.finalImageUID
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
