package reddit_oauth

type OAuthTokenRetriever interface {
	NewOAuthToken() *OAuthToken
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	DeviceID    string `json:"device_id"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}
