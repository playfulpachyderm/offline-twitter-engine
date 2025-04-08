# Use Alpine because it comes with musl by default; point of this build is to staticallly compile libc
# Pinning version 1.21.4 because 1.22 crashes when compiling go-sqlite3 on something in `sqlite3-binding.c`.
from golang:1.23.8-alpine

run apk add --no-cache git sqlite-libs sqlite-dev build-base mingw-w64-gcc curl vim grep

# Install golangci-lint
run curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/go/bin v2.0.2

# Install templ
run GOBIN=/usr/local/go/bin go install github.com/a-h/templ/cmd/templ@v0.3.857
