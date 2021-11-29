#!/bin/bash

set -x
set -e

go build -ldflags="-s -w" -o tw ./twitter
chmod +x tw
