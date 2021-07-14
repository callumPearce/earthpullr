package reddit_cli

import (
	"context"
	"earthpullr/src/config"
	"earthpullr/src/reddit_oauth"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func GetThreadListings(oAuthToken reddit_oauth.OAuthToken, configMan config.ConfigManager) {
	listingsConf, err := configMan.GetMultiConfig([]string{
		"reddit_grant_type_header",
		"reddit_device_id_header",
		"reddit_api_endpoint",
		"reddit_content_type_header",
		"platform",
		"application_name",
		"version",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	userAgent := listingsConf["platform"] + ":" + listingsConf["application_name"] + ":" + listingsConf["version"]

	client := &http.Client{}
	body := url.Values{}
	body.Set("grant_type", listingsConf["reddit_grant_type_header"])
	body.Set("device_id", listingsConf["reddit_device_id_header"])
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		listingsConf["reddit_api_endpoint"]+"/r/EarthPorn/hot",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		panic(err)
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", listingsConf["reddit_content_type_header"])
	req.Header.Add("Authorization", oAuthToken.TokenType+" "+oAuthToken.AccessToken)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(resBody)
	fmt.Println(bodyStr)
}
