package main

import (
	"context"
	"earthpullr/internal/reddit_cli"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
	"earthpullr/pkg/file_readers"
	"earthpullr/pkg/log"
	_ "embed"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"os"
)

//go:embed frontend/build/static/js/main.js
var js string

//go:embed frontend/build/static/css/main.css
var css string

func main() {

	logger := log.New()
	zap.ReplaceGlobals(logger)

	secretsMan := getSecretsManager(logger, "secrets.json")
	configMan := getConfigManager(logger, "config.json")
	ctx := context.Background()
	retriever, err := reddit_cli.NewBackgroundRetriever(ctx, logger, configMan, secretsMan)
	if err != nil {
		logger.Fatal("Failed to create background retriever", zap.Error(err))
		os.Exit(1)
	}

	app := wails.CreateApp(&wails.AppConfig{
		Width:  1024,
		Height: 768,
		Title:  "earthpullr",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})
	app.Bind(retriever)
	app.Run()
}

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
