#!/bin/bash

# BEFORE RUNNING make sure the version number is updated as required in src/internal/config/config.go

wails build -x darwin/amd64 -f -p
wails build -x windows/amd64 -f -p
wails build -x linux/amd64 -f -p

# AFTER RUNNING README.md needs to be manually updated with links to the latest builds
