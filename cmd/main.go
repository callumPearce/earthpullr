package main

import (
	"context"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
	"earthpullr/pkg/file_readers"
	"earthpullr/pkg/log"
	"earthpullr/internal/reddit_cli"
	"fmt"
	"go.uber.org/zap"
	"os"
)

func getSecretsManager(logger *zap.Logger, jsonFilePath string) secrets.SecretsManager {
	flatJsonSecrets, err := file_readers.NewFlatJsonFile(jsonFilePath)
	if err != nil {
		logger.Fatal("Failed to create Secrets Manager", zap.Error(err))
		os.Exit(1)
	}
	secretsMan := secrets.FlatJsonFileSecretManagerAdaptor{
		FlatJsonFile: flatJsonSecrets,
	}
	return secretsMan
}

func getConfigManager(logger *zap.Logger, jsonFilePath string) config.ConfigManager {
	flatJsonConfig, err := file_readers.NewFlatJsonFile(jsonFilePath)
	if err != nil {
		logger.Fatal("Failed to create Config Manager", zap.Error(err))
		os.Exit(1)
	}
	configMan := config.FlatJsonFileConfigManagerAdaptor{
		FlatJsonFile: flatJsonConfig,
	}
	return configMan
}

func main() {
	logger := log.New()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	secretsMan := getSecretsManager(logger, "secrets.json")
	configMan := getConfigManager(logger, "config.json")
	ctx := context.Background()
	retriever, err := reddit_cli.NewBackgroundRetriever(logger, configMan, secretsMan)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to retrieve backgrounds from reddit: %v", err))
	}
	err = retriever.GetBackgrounds(ctx)
	if err != nil {
		logger.Fatal("failed to retrieve backgrounds from reddit", zap.Error(err))
	}
}