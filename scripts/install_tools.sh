echo "Installing golangci..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.52.2

echo "Installing gosec..."
curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.13.1

echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
