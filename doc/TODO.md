# TODO Items

## Frontend
- Change directory path input field to be a directory selector button

## Backend

### Class changes
- background_retriever - `maxAggregatedQueryTimeSecs` needs to be used to limit how long backgrounds are searched for
- reddit_oauth - Retry logic on HTTP request
- reddit_cli - Retry logic on HTTP requests

### Automatically set images as background each OS
- The user should be able to specify if they want earthpullr to automatically set the images downloaded
as their desktop background once they are downloaded. They should be set as a slideshow.
  - Swtich to only support MacOS for now
  - Set mac desktop background using this command: `osascript -e "tell application \"Finder\" to set desktop picture to \"/path/to/image.jpg" as POSIX file"`
  - Provide a button which select the next image along in the downloaded backgrounds to set the background to