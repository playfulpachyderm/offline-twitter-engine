#!/bin/bash

set -x
set -e

if [[ -n "$1" ]]; then
	go build -ldflags="-s -w -X main.version_string=$1 -X webserver.is_production=true" -o tw ./twitter
else
	go build -ldflags="-s -w -X webserver.is_production=true" -o tw ./twitter
fi
chmod +x tw
