package reddit_oauth

import (
	"context"
	"earthpullr/internal/secrets"
	"earthpullr/pkg/config"
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
	redditAppClientID string
	oauthConf         map[string]string
	userAgent         string
	client            *http.Client
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequestBody() string {
	oAuthBody := url.Values{}
	oAuthBody.Set("grant_type", oAuthRequest.oauthConf["reddit_grant_type_header"])
	oAuthBody.Set("device_id", oAuthRequest.oauthConf["reddit_device_id_header"])
	return oAuthBody.Encode()
}

func (oAuthRequest *ApplicationOnlyOAuthRequest) getPostRequest(ctx context.Context) (*http.Request, error) {
	oAuthBody := oAuthRequest.getPostRequestBody()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		oAuthRequest.oauthConf["reddit_access_token_url"],
		strings.NewReader(oAuthBody),
	)
	req.Header.Add("Content-Type", oAuthRequest.oauthConf["reddit_content_type_header"])
	req.Header.Add("User-Agent", oAuthRequest.userAgent)
	req.SetBasicAuth(oAuthRequest.redditAppClientID, "")
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

func NewApplicationOnlyOAuthRequest(ctx context.Context, client *http.Client, secretMan secrets.SecretsManager, configMan config.ConfigManager) (appOnlyOAuthReq ApplicationOnlyOAuthRequest, err error) {
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
		oauthConf:         oauthConf,
		userAgent:         oauthConf["platform"] + ":" + oauthConf["application_name"] + ":" + oauthConf["version"],
		redditAppClientID: client_id,
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
