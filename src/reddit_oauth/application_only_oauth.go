package reddit_oauth

import (
	"context"
	"earthpullr/src/config"
	"earthpullr/src/secrets"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ApplicationOnlyOAuthRequest struct {
	RequestURL        string
	GrantType         string
	DeviceID          string
	ContentType       string
	UserAgent         string
	RedditAppClientID string
	Client            *http.Client
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequestBody() string {
	oAuthBody := url.Values{}
	oAuthBody.Set("grant_type", oAuthRequest.GrantType)
	oAuthBody.Set("device_id", oAuthRequest.DeviceID)
	return oAuthBody.Encode()
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequest() (*http.Request, error) {
	oAuthBody := oAuthRequest.getPostRequestBody()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		oAuthRequest.RequestURL,
		strings.NewReader(oAuthBody),
	)
	req.Header.Add("Content-Type", oAuthRequest.ContentType)
	req.Header.Add("User-Agent", oAuthRequest.UserAgent)
	req.SetBasicAuth(oAuthRequest.RedditAppClientID, "")
	return req, err
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) sendPostRequest(req *http.Request) (*http.Response, error) {
	res, err := oAuthRequest.Client.Do(req)
	return res, err
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) extractResponse(response *http.Response) (oAuthToken OAuthToken, err error) {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	bodyStr := string(body)
	oAuthToken = OAuthToken{}
	json.Unmarshal([]byte(bodyStr), &oAuthToken)
	if response.StatusCode == http.StatusTooManyRequests {
		panic(fmt.Sprintf("Rate Limited status: got %v", response.Status))
	} else if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Error status: got %v", response.Status))
	}
	return oAuthToken, err
}

func (oAuthRequest ApplicationOnlyOAuthRequest) NewOAuthToken() OAuthToken {
	req, err := oAuthRequest.getPostRequest()
	if err != nil {
		panic(err)
	}
	res, err := oAuthRequest.sendPostRequest(req)
	if err != nil {
		panic(err)
	}
	oAuthToken, err := oAuthRequest.extractResponse(res)
	if err != nil {
		panic(err)
	}
	return oAuthToken
}

func NewApplicationOnlyOAuthRequest(secretMan secrets.SecretsManager, configMan config.ConfigManager) (appOnlyOAuthReq ApplicationOnlyOAuthRequest, err error) {
	errStrPrefix := "failed to create http request to retrieve application only oauth auth token: "
	oauthConf, err := configMan.GetMultiConfig([]string{
		"reddit_grant_type_header",
		"reddit_access_token_url",
		"reddit_device_id_header",
		"reddit_content_type_header",
		"platform",
		"application_name",
		"version",
	})
	if err != nil {
		return appOnlyOAuthReq, fmt.Errorf(errStrPrefix+"%v", err)
	}
	client_id, err := secretMan.GetSecret("reddit_app_client_id")
	if err != nil {
		return appOnlyOAuthReq, fmt.Errorf(errStrPrefix+"%v", err)
	}
	appOnlyOAuthReq = ApplicationOnlyOAuthRequest{
		RequestURL:        oauthConf["reddit_access_token_url"],
		GrantType:         oauthConf["reddit_grant_type_header"],
		DeviceID:          oauthConf["reddit_device_id_header"],
		ContentType:       oauthConf["reddit_content_type_header"],
		UserAgent:         oauthConf["platform"] + ":" + oauthConf["application_name"] + ":" + oauthConf["version"],
		RedditAppClientID: client_id,
		Client:            &http.Client{},
	}
	return appOnlyOAuthReq, err
}
