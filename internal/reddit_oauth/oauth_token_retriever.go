package reddit_oauth

import "context"

type OAuthTokenRetriever interface {
	NewOAuthToken() *OAuthToken
	FromContext(ctx context.Context) (*OAuthToken, error)
	ToContext(ctx context.Context, token OAuthToken) context.Context
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	DeviceID    string `json:"device_id"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}