package reddit_oauth

import (
	"context"
	"earthpullr/internal/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const oAuthTokenKey int = 0

type ApplicationOnlyOAuthRequest struct {
	request           *http.Request
	conf         	  config.Config
	userAgent         string
	client            *http.Client
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequestBody() string {
	oAuthBody := url.Values{}
	oAuthBody.Set("grant_type", oAuthRequest.conf.RedditGrantTypeHeader)
	oAuthBody.Set("device_id", oAuthRequest.conf.RedditDeviceIdHeader)
	return oAuthBody.Encode()
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequest(ctx context.Context) (*http.Request, error) {
	oAuthBody := oAuthRequest.getPostRequestBody()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		oAuthRequest.conf.RedditAccessTokenUrl,
		strings.NewReader(oAuthBody),
	)
	req.Header.Add("Content-Type", oAuthRequest.conf.RedditContentTypeHeader)
	req.Header.Add("User-Agent", oAuthRequest.userAgent)
	req.SetBasicAuth(oAuthRequest.conf.RedditAppClientId, "")
	return req, err
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) doRequest() (*http.Response, error) {
	res, err := oAuthRequest.client.Do(oAuthRequest.request)
	return res, err
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) extractResponse(response *http.Response) (oAuthToken OAuthToken, err error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read oauth request response body: %v", err)
		return oAuthToken, err
	}

	bodyStr := string(body)
	oAuthToken = OAuthToken{}
	err = json.Unmarshal([]byte(bodyStr), &oAuthToken)
	if err != nil {
		err = fmt.Errorf("failed to parse oauth request response body json: %v", err)
		return oAuthToken, err
	}

	if response.StatusCode == http.StatusTooManyRequests {
		err = fmt.Errorf("rate limited status: got %v", response.Status)
	} else if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("error status: got %v", response.Status)
	}
	return oAuthToken, err
}

func (oAuthRequest ApplicationOnlyOAuthRequest) NewOAuthToken() (oAuthTokenPtr *OAuthToken, err error) {
	res, err := oAuthRequest.doRequest()
	if err != nil {
		return oAuthTokenPtr, err
	}
	oAuthToken, err := oAuthRequest.extractResponse(res)
	if err != nil {
		return oAuthTokenPtr, err
	}
	oAuthTokenPtr = &oAuthToken
	return &oAuthToken, err
}

func NewApplicationOnlyOAuthRequest(ctx context.Context, client *http.Client, conf config.Config) (appOnlyOAuthReq ApplicationOnlyOAuthRequest, err error) {
	appOnlyOAuthReq = ApplicationOnlyOAuthRequest{
		conf:         conf,
		userAgent:         conf.Platform + ":" + conf.ApplicationName + ":" + conf.Version,
		client:            client,
	}
	req, err := appOnlyOAuthReq.getPostRequest(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create an application only oauth request: %v", err)
		return appOnlyOAuthReq, err
	}
	appOnlyOAuthReq.request = req

	return appOnlyOAuthReq, err
}

func FromContext(ctx context.Context) (*OAuthToken, error) {
	if ctx == nil {
		return &OAuthToken{}, fmt.Errorf("cannot retrieve OAuthToken from nil context")
	}
	oAuthToken, ok := ctx.Value(oAuthTokenKey).(*OAuthToken)
	if !ok {
		return &OAuthToken{}, fmt.Errorf("oauth token is not in the current context")
	}
	return oAuthToken, nil
}

func ToContext(ctx context.Context, token *OAuthToken) context.Context {
	return context.WithValue(ctx, oAuthTokenKey, token)
}
