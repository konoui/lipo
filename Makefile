VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
SRC_DIR := ./
BIN_NAME := lipo
BINARY := bin/$(BIN_NAME)

GOLANGCI_LINT_VERSION := v1.45.2
export GO111MODULE=on

## Build binaries on your environment
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(SRC_DIR)

lint:
	@(if ! type golangci-lint >/dev/null 2>&1; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION) ;fi)
	golangci-lint run ./...

test:
	go test -v ./...

clean:
	rm -f $(BIN_NAME)
