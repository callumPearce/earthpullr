package reddit_cli

import (
	"earthpullr/src/config"
	"earthpullr/src/secrets"
	"fmt"
	"strconv"
)

type BackgroundRetriever struct {
	configMan                  config.ConfigManager
	secretsMan                 secrets.SecretsManager
	width                      int
	height                     int
	subreddit                  string
	subredditSearchType        string
	queryBatchSize             int
	maxAggregatedQueryTimeSecs int
	backgroundsCount           int
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
	queryBatchSize, err := strconv.Atoi(backgroundConf["query_batch_size"])
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
		subreddit:                  backgroundConf["subreddit"],
		subredditSearchType:        backgroundConf["subreddit_search_type"],
		width:                      width,
		height:                     height,
		queryBatchSize:             queryBatchSize,
		maxAggregatedQueryTimeSecs: maxAggregatedQueryTimeSecs,
		backgroundsCount:           backgroundsCount,
	}
	return retriever, nil
}
