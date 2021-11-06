# TODO Items

### Frontend

### bugs

### background_retriever
- Should not count an image that has already been downloaded as a new image 
  - Before downloading create a map of existing images downloaded by anaylsing directory, do not download
    any files which would end-up having the same the same as what is stored in this map
  - Could instead store all images downloaded into a config file which gets read into a map on every pull request. Store
    this in the directory where all the images stored

### reddit_oauth
- Retry logic on HTTP request

### reddit_cli
- Retry logic on HTTP requests

### secrets
- Move to reddit oauth client ID to config, it's not really a secret?