#!/bin/bash

set -x
set -e

export CGO_ENABLED=1
export CC=musl-gcc

FLAGS="-s -w -linkmode=external -extldflags=-static -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true"

if [[ -n "$1" ]]; then
	go build -ldflags="$FLAGS -X main.version_string=$1" -o tw ./twitter
else
	go build -ldflags="$FLAGS" -o tw ./twitter
fi
chmod +x tw
