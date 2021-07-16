package main

import (
	"earthpullr/src/config"
	"earthpullr/src/file_readers"
	"earthpullr/src/reddit_cli"
	"earthpullr/src/reddit_oauth"
	"earthpullr/src/secrets"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func getSecretsManager(jsonFilePath string) secrets.SecretsManager {
	flatJsonSecrets, err := file_readers.NewFlatJsonFile(jsonFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create Secrets Manager: %v", err))
		os.Exit(1)
	}
	secretsMan := secrets.FlatJsonFileSecretManagerAdaptor{
		FlatJsonFile: flatJsonSecrets,
	}
	return secretsMan
}

func getConfigManager(jsonFilePath string) config.ConfigManager {
	flatJsonConfig, err := file_readers.NewFlatJsonFile(jsonFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create Config Manager: %v", err))
		os.Exit(1)
	}
	configMan := config.FlatJsonFileConfigManagerAdaptor{
		FlatJsonFile: flatJsonConfig,
	}
	return configMan
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

func getListingResponse(oauthToken *reddit_oauth.OAuthToken, client *http.Client, cm config.ConfigManager) reddit_cli.ListingResponse {
	listingParams := reddit_cli.ListingParameters{
		Subreddit:    "earthporn",
		ListingLimit: 10,
		SearchType:   "new",
		Before:       "",
	}
	listingRequest, err := reddit_cli.NewListingRequest(
		client,
		oauthToken,
		cm,
		listingParams,
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to get Listings for EarthPorn subreddit: %v", err))
		os.Exit(1)
	}
	listingResponse, err := listingRequest.DoRequest()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to get Listings for EarthPorn subreddit: %v", err))
		os.Exit(1)
	}
	return listingResponse
}

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	secretsMan := getSecretsManager("secrets.json")
	configMan := getConfigManager("config.json")
	oauthToken := getOAuthToken(client, secretsMan, configMan)
	listingResponse := getListingResponse(oauthToken, client, configMan)
	imagesRetriever, err := reddit_cli.NewImagesRetriever(listingResponse, oauthToken, client)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed retriever image batch: %v", err))
	}
	imagesRetriever.SaveImages("images")
}
