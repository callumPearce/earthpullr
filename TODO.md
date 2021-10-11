# TODO Items

### Frontend

### bugs

### background_retriever
- Should store the latest image id retrieved (for a certain period of time) and only get images after that
- Should not count an image that has already been downloaded as a new image (solve with the one above)

### reddit_oauth
- Retry logic on HTTP request
- Background retriever should check if token is valid rather than retrieving a new one on every request

### reddit_cli
- Retry logic on HTTP requests

### secrets
- Can't be released until secret key is not stored in plaintext on machine 