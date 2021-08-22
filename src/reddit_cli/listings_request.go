package reddit_cli

import (
	"context"
	"earthpullr/src/config"
	"earthpullr/src/reddit_oauth"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ListingRequest struct {
	listingsConf map[string]string
	client       *http.Client
	oAuthToken   *reddit_oauth.OAuthToken
	request      *http.Request
	before       string
	after        string
}

type ListingResponse struct {
	Kind string      `json:"kind"`
	Data listingData `json:"data"`
}

type listingData struct {
	Children []listingChild `json:"children"`
}

type listingChild struct {
	Data listingChildData `json:"data"`
}

type listingChildData struct {
	Title   string             `json:"title"`
	Preview imagePreviewParent `json:"preview"`
	Name    string             `json:"name"`
}

type imagePreviewParent struct {
	ImagesList []image `json:"images"`
}

type image struct {
	Source sourceImage `json:"source"`
}

type sourceImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (lr *ListingRequest) setListingsConfig(configMan config.ConfigManager) error {
	listingsConf, err := configMan.GetMultiConfig([]string{
		"reddit_grant_type_header",
		"reddit_device_id_header",
		"reddit_api_endpoint",
		"reddit_content_type_header",
		"platform",
		"application_name",
		"version",
		"subreddit",
		"subreddit_search_type",
		"query_batch_size",
	})
	if err != nil {
		return err
	}
	lr.listingsConf = listingsConf
	return err
}

func (lr *ListingRequest) getRequestBody() string {
	body := url.Values{}
	body.Set("grant_type", lr.listingsConf["reddit_grant_type_header"])
	body.Set("device_id", lr.listingsConf["reddit_device_id_header"])
	return body.Encode()
}

func (lr *ListingRequest) setRequestHeaders(ctx context.Context, req *http.Request) error {
	req.Header.Add("User-Agent", lr.listingsConf["platform"]+":"+lr.listingsConf["application_name"]+":"+lr.listingsConf["version"])
	req.Header.Add("Content-Type", lr.listingsConf["reddit_content_type_header"])
	oAuthToken, err := reddit_oauth.FromContext(ctx)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", oAuthToken.TokenType+" "+oAuthToken.AccessToken)
	return nil
}

func (lr *ListingRequest) setRequestQueryParams(req *http.Request) {
	q := req.URL.Query()
	q.Add("limit", lr.listingsConf["query_batch_size"])
	if lr.before != "" {
		q.Add("after", lr.before)
	}
	if lr.after != "" {
		q.Add("after", lr.after)
	}
	req.URL.RawQuery = q.Encode()
}

func (lr *ListingRequest) getRequest(ctx context.Context) (*http.Request, error) {
	body := lr.getRequestBody()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		lr.listingsConf["reddit_api_endpoint"]+"/r/"+lr.listingsConf["subreddit"]+"/"+lr.listingsConf["subreddit_search_type"],
		strings.NewReader(body),
	)
	if err != nil {
		err = fmt.Errorf("request to retrieve reddit listings could not be created: %v", err)
		return req, err
	}
	lr.setRequestQueryParams(req)
	err = lr.setRequestHeaders(ctx, req)
	return req, err
}

func (lr ListingRequest) DoRequest() (lres ListingResponse, err error) {
	res, err := lr.client.Do(lr.request)
	if err != nil {
		return lres, err
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("failed to read listing request response body: %v", err)
		return lres, err
	}

	bodyStr := string(resBody)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		err = fmt.Errorf(
			"get request to retrieve listings returned status code %d, full body response: %s",
			res.StatusCode,
			bodyStr,
		)
	} else {
		err = json.Unmarshal([]byte(bodyStr), &lres)
		if err != nil {
			err = fmt.Errorf("failed to parse listings request response from reddit: %v", err)
		}
	}
	return lres, err
}

	func NewListingRequest(
	ctx context.Context,
	client *http.Client,
	confMan config.ConfigManager,
	before string,
	after string,
) (lr ListingRequest, err error) {
	err = lr.setListingsConfig(confMan)
	if err != nil {
		return lr, fmt.Errorf("failed to create listings request due to Config variables - %v", err)
	}
	lr.client = client
	lr.before = before
	lr.after = after
	req, err := lr.getRequest(ctx)
	lr.request = req
	if err != nil {
		return lr, fmt.Errorf("failed to create listings request - %v", err)
	}
	return lr, err
}
