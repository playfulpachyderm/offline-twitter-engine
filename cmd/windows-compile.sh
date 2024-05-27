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

go build -ldflags="-linkmode external -extldflags -static -s -w -X main.version_string=$1 -X gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver.use_embedded=true" -o twitter.exe ./twitter
