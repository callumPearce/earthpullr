package main

import (
	"earthpullr/src/reddit_oauth"
	"earthpullr/src/secrets"
	"fmt"
)

func main() {
	secrets_man := secrets.NewJsonFileSecrets("../secrets.json")
	var redditOauth reddit_oauth.OAuthTokenRetriever = reddit_oauth.NewApplicationOnlyOAuthRequest(secrets_man)
	var oauthToken reddit_oauth.OAuthToken = redditOauth.NewOAuthToken()
	fmt.Println(oauthToken.AccessToken)
}
