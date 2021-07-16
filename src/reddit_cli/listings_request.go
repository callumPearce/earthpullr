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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ListingRequest struct {
	listingsParameters ListingParameters
	listingsConf       map[string]string
	client             *http.Client
	oAuthToken         *reddit_oauth.OAuthToken
	request            *http.Request
}

type ListingParameters struct {
	Subreddit    string
	ListingLimit int
	SearchType   string
	Before       string
	After        string
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

func (lr *ListingRequest) setRequestHeaders(req *http.Request) {
	req.Header.Add("User-Agent", lr.listingsConf["platform"]+":"+lr.listingsConf["application_name"]+":"+lr.listingsConf["version"])
	req.Header.Add("Content-Type", lr.listingsConf["reddit_content_type_header"])
	req.Header.Add("Authorization", lr.oAuthToken.TokenType+" "+lr.oAuthToken.AccessToken)
}

func (lr *ListingRequest) setRequestQueryParams(req *http.Request) {
	q := req.URL.Query()
	q.Add("limit", strconv.Itoa(lr.listingsParameters.ListingLimit))
	if lr.listingsParameters.Before != "" {
		q.Add("after", lr.listingsParameters.Before)
	}
	if lr.listingsParameters.After != "" {
		q.Add("after", lr.listingsParameters.After)
	}
	req.URL.RawQuery = q.Encode()
}

func (lr *ListingRequest) getRequest() (*http.Request, error) {
	body := lr.getRequestBody()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		lr.listingsConf["reddit_api_endpoint"]+"/r/"+lr.listingsParameters.Subreddit+"/"+lr.listingsParameters.SearchType,
		strings.NewReader(body),
	)
	if err != nil {
		err = fmt.Errorf("request to retrieve reddit listings could not be created: %v", err)
		return req, err
	}
	lr.setRequestQueryParams(req)
	lr.setRequestHeaders(req)
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
	client *http.Client,
	oAuthToken *reddit_oauth.OAuthToken,
	confMan config.ConfigManager,
	listingParameters ListingParameters,
) (lr ListingRequest, err error) {
	err = lr.setListingsConfig(confMan)
	if err != nil {
		log.Error("Failed to create listings request due to Config variables - %v", err)
		return lr, err
	}
	lr.listingsParameters = listingParameters
	lr.client = client
	lr.oAuthToken = oAuthToken
	req, err := lr.getRequest()
	lr.request = req
	if err != nil {
		log.Error("Failed to create listings request - %v", err)
		return lr, err
	}
	return lr, err
}

func NewListingParameters(subreddit string, listingLimit int, searchType string) (lr ListingParameters, err error) {
	if listingLimit > 100 {
		return lr, fmt.Errorf("listimingLimit exceeded: %d > 100", listingLimit)
	}

	searchTypes := []string{
		"hot",
		"new",
		"random",
		"rising",
		"top",
	}
	found := false
	for _, stype := range searchTypes {
		if stype == searchType {
			found = true
			break
		}
	}
	if !found {
		return lr, fmt.Errorf("unknown search type '%s'", searchType)
	}

	lr.ListingLimit = listingLimit
	lr.SearchType = searchType
	lr.Subreddit = subreddit
	return lr, err
}

func (lr *ListingParameters) WithBefore(before string) error {
	if lr.After != "" {
		return fmt.Errorf("cannot set 'before', 'after' parameter has already been set for this group listings parameters")
	}
	lr.Before = before
	return nil
}

func (lr *ListingParameters) WithAfter(after string) error {
	if lr.Before != "" {
		return fmt.Errorf("cannot set 'after', 'before' parameter has already been set for this group listings parameters")
	}
	lr.After = after
	return nil
}
