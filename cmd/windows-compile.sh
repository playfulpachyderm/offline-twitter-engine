#!/bin/sh

set -x
set -e

export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export GOOS=windows
export GOARCH=amd64

if [[ -z "$1" ]]; then
	echo "No version number given!  Exiting..."
	exit 1
fi

# Always use static build for windows
FLAGS="-s -w -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true"
SPECIAL_FLAGS_FOR_STATIC_BUILD="-linkmode=external -extldflags=-static"
go build -ldflags="$FLAGS $SPECIAL_FLAGS_FOR_STATIC_BUILD -X main.version_string=$1" -o twitter.exe ./twitter
