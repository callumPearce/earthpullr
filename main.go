package main

import (
	"context"
	"earthpullr/internal/reddit_cli"
	"earthpullr/internal/config"
	"earthpullr/pkg/log"
	_ "embed"
	"github.com/kbinani/screenshot"
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

	conf := getConfig(logger, "")
	ctx := context.Background()
	retriever, err := reddit_cli.NewBackgroundRetriever(ctx, logger, conf)
	if err != nil {
		logger.Fatal("Failed to create background retriever", zap.Error(err))
		os.Exit(1)
	}

	bounds := screenshot.GetDisplayBounds(0)
	width := bounds.Dx()/5
	height := int(float64(width) * 1.15)

	app := wails.CreateApp(&wails.AppConfig{
		MaxWidth:  width,
		MaxHeight: height,
		Title:  "earthpullr",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})
	app.Bind(retriever)
	app.Run()
}

func getConfig(logger *zap.Logger, jsonFilePath string) config.Config {
	conf, err := config.NewConfig(jsonFilePath)
	if err != nil {
		logger.Fatal("Failed to read config file, shutting down.", zap.Error(err))
		os.Exit(1)
	}
	return conf
}
