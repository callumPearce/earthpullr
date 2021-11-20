#!/bin/bash

# BEFORE RUNNING make sure the version number is updated as required in src/internal/config/config.go

wails build -f -p  # -x darwin/amd64 Disabled cross compilation as it's failing for mac
#wails build -x windows/amd64 -f -p
#wails build -x linux/amd64 -f -p

# AFTER RUNNING README.md needs to be manually updated with links to the latest builds
