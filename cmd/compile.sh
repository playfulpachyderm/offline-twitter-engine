#!/bin/bash

set -x
set -e

# General build flags
FLAGS="-s -w -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true"

if [[ -n "$1" ]]; then
	# Static build params
	export CGO_ENABLED=1
	export CC=musl-gcc
	SPECIAL_FLAGS_FOR_STATIC_BUILD="-linkmode=external -extldflags=-static"
	go build -ldflags="$FLAGS $SPECIAL_FLAGS_FOR_STATIC_BUILD -X main.version_string=$1" -o tw ./twitter
else
	go build -ldflags="$FLAGS" -o tw ./twitter
fi
chmod +x tw
