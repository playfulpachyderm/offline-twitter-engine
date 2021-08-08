#!/bin/bash

set -x
set -e

go build -o tw ./twitter
chmod +x tw
