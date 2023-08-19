#!/bin/bash

set -x
set -e

if [[ -n "$1" ]]; then
	go build -ldflags="-s -w -X main.version_string=$1 -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true" -o tw ./twitter
else
	go build -ldflags="-s -w -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true" -o tw ./twitter
fi
chmod +x tw
