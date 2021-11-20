package reddit_cli

import (
	"context"
	"earthpullr/internal/config"
	"earthpullr/internal/reddit_oauth"
	"earthpullr/internal/user_settings"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/wailsapp/wails"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type BackgroundRetriever struct {
	logger                     *zap.Logger
	conf                       config.Config
	runtime                    *wails.Runtime
	ctx                        context.Context
	client                     *http.Client
	userSettingsMan            user_settings.UserSettingsManager
}

type BackgroundsRequest struct {
	Width          int
	Height         int
	BackgroundsCount int
	DownloadPath   string
}

func NewBackgroundRetriever(ctx context.Context, logger *zap.Logger, conf config.Config) (*BackgroundRetriever, error) {
	userSettingsMan, err := user_settings.NewUserSettingsManager(conf.UserSettingsFname)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saved settings: %v", err)
	}
	retriever := &BackgroundRetriever{
		logger:                     logger,
		conf:                       conf,
		ctx:                        ctx,
		client:                     &http.Client{Timeout: 10 * time.Second},
		userSettingsMan: 				userSettingsMan,
	}
	return retriever, nil
}

func (br *BackgroundRetriever) WailsInit(runtime *wails.Runtime) error {
	br.runtime = runtime
	return nil
}

func (br *BackgroundRetriever) GetUserDownloadPath() string {
	br.logger.Debug("Retrieved download path: " + br.userSettingsMan.Settings.DownloadPath)
	return br.userSettingsMan.Settings.DownloadPath
}

func (br *BackgroundRetriever) GetBackgrounds(request map[string]interface{}) (string, error) {
	var brRequest BackgroundsRequest
	err := mapstructure.Decode(request, &brRequest)
	if err != nil {
		return "Error", fmt.Errorf("failed to decode backgrounds request from frontend: %v", err)
	}
	if _, dirErr := os.Stat(brRequest.DownloadPath); os.IsNotExist(dirErr) {
		return "Error", fmt.Errorf("Download path '%s' does not exist", brRequest.DownloadPath)
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
	existingBackgrounds := NewExistingBackgrounds(brRequest.DownloadPath, br.conf.ExistingImagesFilename, br.logger)
	err = br.getBackgroundsWithBatching(brRequest, existingBackgrounds)
	if err != nil {
		return "Error", err
	}
	err = br.userSettingsMan.SaveNewUserSettings(brRequest.DownloadPath)
	if err != nil {
		br.logger.Error("Failed to save user settings", zap.Error(err))
		return "Error", fmt.Errorf("failed to save user settings")
	}
	return "Success", nil
}

func (br *BackgroundRetriever) getBackgroundsWithBatching(brRequest BackgroundsRequest, existingBackgrounds *ExistingBackgrounds) error {
	savedImages := 0
	afterUID := ""
	for savedImages < brRequest.BackgroundsCount {
		listingRequest, err := NewListingRequest(
			br.ctx,
			br.client,
			br.conf,
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
		imagesRetriever, err := NewImagesRetriever(br.logger, br.ctx, listingResponse, br.client, remainingImagesCount, brRequest.Width, brRequest.Height, existingBackgrounds)
		afterUID = imagesRetriever.finalImageUID
		if err != nil {
			err = fmt.Errorf("failed to retrieve image batch: %v", err)
			return err
		}
		imagesRetriever.SaveImages(brRequest.DownloadPath, br.runtime, existingBackgrounds)
		savedImages += imagesRetriever.imageCount
	}
	return existingBackgrounds.SaveExistingBackgrounds()
}

func (br *BackgroundRetriever) addOAuthTokenToCtx() error {
	redditOauth, err := reddit_oauth.NewApplicationOnlyOAuthRequest(br.ctx, br.client, br.conf)
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
