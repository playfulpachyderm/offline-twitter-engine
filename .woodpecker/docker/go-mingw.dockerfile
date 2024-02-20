from golang:alpine

run apk add --no-cache git sqlite-libs sqlite-dev build-base mingw-w64-gcc curl vim

# Install golangci-lint
run curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/go/bin v1.53.1

# Install project dependencies (so they don't have to be reinstalled on every CI run)
run git clone https://gitlab.com/offline-twitter/twitter_offline_engine.git && cd twitter_offline_engine && go install ./... && cd .. && rm -r twitter_offline_engine
