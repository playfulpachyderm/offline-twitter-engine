from ubuntu:jammy

run apt update && apt install -y sudo curl wget build-essential sqlite3 jq git musl-dev musl-tools

# Install go and golangci-lint
run wget https://go.dev/dl/go1.23.8.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.23.8.linux-amd64.tar.gz
env PATH="$PATH:/usr/local/go/bin"
run curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/go/bin v2.0.2

# Install templ
run GOBIN=/usr/local/go/bin go install github.com/a-h/templ/cmd/templ@v0.3.857

# For SSH upload
copy known_hosts /root/.ssh/known_hosts

# Install NodeJS v20
run curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs
