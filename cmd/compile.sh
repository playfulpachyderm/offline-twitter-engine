#!/bin/bash

set -x
set -e

# General build flags
FLAGS="-s -w -X gitlab.com/offline-twitter/twitter_offline_engine/pkg/webserver.use_embedded=true"

# Check for the `--static` flag and consume it
USE_STATIC=false
if [[ "$1" == "--static" ]] || [[ "$1" == "-static" ]]; then
	USE_STATIC=true
	shift
fi

if [[ -z "$1" ]]; then
	# If no version string, it's a development build
	go build -ldflags="$FLAGS" -o tw ./twitter
elif $USE_STATIC; then
	# Static build params
	export CGO_ENABLED=1
	export CC=musl-gcc
	SPECIAL_FLAGS_FOR_STATIC_BUILD="-linkmode=external -extldflags=-static"
	go build -ldflags="$FLAGS $SPECIAL_FLAGS_FOR_STATIC_BUILD -X main.version_string=$1" -o tw ./twitter
else
	# Version string, but not static
	go build -ldflags="$FLAGS -X main.version_string=$1" -o tw ./twitter
fi

chmod +x tw
