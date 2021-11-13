package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	RedditAccessTokenUrl       string `json:"reddit_access_token_url"`
	RedditGrantTypeHeader      string `json:"reddit_grant_type_header"`
	RedditDeviceIdHeader       string `json:"reddit_device_id_header"`
	RedditContentTypeHeader    string `json:"reddit_content_type_header"`
	RedditApiEndpoint          string `json:"reddit_api_endpoint"`
	Version                    string `json:"version"`
	Platform                   string `json:"platform"`
	ApplicationName            string `json:"application_name"`
	Subreddit                  string `json:"subreddit"`
	SubredditSearchType        string `json:"subreddit_search_type"`
	QueryBatchSize             int    `json:"query_batch_size"`
	MaxAggregatedQueryTimeSecs int    `json:"max_aggregated_query_time_secs"`
	ExistingImagesFilename     string `json:"existing_images_filename"`
	RedditAppClientId          string `json:"reddit_app_client_id"`
	UserSettingsFname          string `json:"user_settings_fname"`
}

func NewConfig(fpathOverride string) (Config, error) {
	if fpathOverride != "" {
		return NewConfigFromFile(fpathOverride)
	} else {
		return NewDefaultConfig(), nil
	}
}

func NewDefaultConfig() Config {
	return Config{
		RedditAccessTokenUrl: "https://www.reddit.com/api/v1/access_token",
		RedditGrantTypeHeader: "https://oauth.reddit.com/grants/installed_client",
		RedditDeviceIdHeader: "DO_NOT_TRACK_THIS_DEVICE",
		RedditContentTypeHeader: "application/x-www-form-urlencoded",
		RedditApiEndpoint: "https://oauth.reddit.com",
		Version: "v0.0.1",
		Platform: "macOS",
		ApplicationName: "earthpullr",
		Subreddit: "earthporn",
		SubredditSearchType: "hot",
		QueryBatchSize: 100,
		MaxAggregatedQueryTimeSecs: 30,
		ExistingImagesFilename: "earthpullr_existing_images.json",
		RedditAppClientId: "3gMaLS0rRxDTdEWErlrTEg",
		UserSettingsFname: "earthpullr_user_settings.json",
	}
}

func NewConfigFromFile(fpathOverride string) (Config, error) {
	file, err := os.Open(fpathOverride)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshall json config file: %v", err)
	}
	return config, nil
}