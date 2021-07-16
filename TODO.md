# TODO Items

- Shared context for API requests needs to be developed

### reddit_oauth

- Save token locally and check if it is still valid when program restarts
- Implement oauth token refresh query
- Retry logic on HTTP request


### reddit_cli

- Retry logic on HTTP requests
- Detection if oauth token is invalid, in which case refresh it


### secrets

- Cant be released until secret key is not stored in plaintext on machine 