package main

import (
	"earthpullr/src/config"
	"earthpullr/src/file_readers"
	"earthpullr/src/reddit_cli"
	"earthpullr/src/reddit_oauth"
	"earthpullr/src/secrets"
	"fmt"
	"os"

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

func main() {
	secretsMan := getSecretsManager("secrets.json")
	configMan := getConfigManager("config.json")
	var redditOauth reddit_oauth.OAuthTokenRetriever
	redditOauth, err := reddit_oauth.NewApplicationOnlyOAuthRequest(secretsMan, configMan)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to retrieve oauth Token from reddit: %v", err))
		os.Exit(1)
	}
	var oauthToken reddit_oauth.OAuthToken = redditOauth.NewOAuthToken()
	reddit_cli.GetThreadListings(oauthToken, configMan)
}
