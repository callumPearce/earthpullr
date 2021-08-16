package main

import (
	"earthpullr/src/config"
	"earthpullr/src/file_readers"
	"earthpullr/src/reddit_cli"
	"earthpullr/src/secrets"
	"fmt"
	"os"
	"context"

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
	ctx := context.Background()
	retriever, err := reddit_cli.NewBackgroundRetriever(configMan, secretsMan)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to retrieve backgrounds from reddit: %v", err))
	}
	retriever.GetBackgrounds(ctx)
}
