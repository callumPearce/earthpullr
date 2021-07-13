package reddit_oauth

import (
	"context"
	"earthpullr/src/secrets"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	REQUEST_URL  = "https://www.reddit.com/api/v1/access_token"
	GRANT_TYPE   = "https://oauth.reddit.com/grants/installed_client"
	DEVICE_ID    = "DO_NOT_TRACK_THIS_DEVICE"
	CONTENT_TYPE = "application/x-www-form-urlencoded"
	VERSION      = "v1.0.0"
	PLATFORM     = "windows"
	USER_AGENT   = PLATFORM + ":earthpullr:" + VERSION
)

type ApplicationOnlyOAuthRequest struct {
	RequestURL     string
	GrantType      string
	DeviceID       string
	ContentType    string
	UserAgent      string
	Client         *http.Client
	SecretsManager secrets.SecretsManager
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
	req.Header.Add("Content-Type", CONTENT_TYPE)
	req.Header.Add("User-Agent", USER_AGENT)
	req.SetBasicAuth(oAuthRequest.SecretsManager.GetSecret("reddit_app_client_id"), "")
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

func NewApplicationOnlyOAuthRequest(secret_manager secrets.SecretsManager) ApplicationOnlyOAuthRequest {
	appOnlyOAuthReq := ApplicationOnlyOAuthRequest{
		RequestURL:     REQUEST_URL,
		GrantType:      GRANT_TYPE,
		DeviceID:       DEVICE_ID,
		ContentType:    CONTENT_TYPE,
		UserAgent:      USER_AGENT,
		Client:         &http.Client{},
		SecretsManager: secret_manager,
	}
	return appOnlyOAuthReq
}
