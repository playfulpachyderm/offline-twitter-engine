from ubuntu:jammy

run apt update && apt install -y sudo curl wget build-essential sqlite3 jq git

# Install go and golangci-lint
run wget https://go.dev/dl/go1.20.4.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
env PATH="$PATH:/usr/local/go/bin"
run curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/go/bin v1.53.1

# Install project dependencies (so they don't have to be reinstalled on every CI run)
run git clone https://gitlab.com/offline-twitter/twitter_offline_engine.git && cd twitter_offline_engine && go install ./... && cd .. && rm -r twitter_offline_engine
