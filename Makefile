VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
SRC_DIR := ./
BIN_NAME := lipo
BINARY := bin/$(BIN_NAME)

GOLANGCI_LINT_VERSION := v1.62.2
export GO111MODULE=on

CMD_PACKAGE := github.com/konoui/lipo/cmd
LDFLAGS := -X '$(CMD_PACKAGE).Version=$(VERSION)' -X '$(CMD_PACKAGE).Revision=$(REVISION)'

## Build binaries on your environment
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(SRC_DIR)

lint:
	@(if ! type golangci-lint >/dev/null 2>&1; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION) ;fi)
	golangci-lint run ./...

test:
	go test -v ./...

test-large-file:
	./test-large-file.sh

test-on-non-macos:
	./test-on-non-macos.sh

release-test:
	goreleaser --snapshot --skip-publish --clean

cover:
	go test -coverpkg=./... -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

clean:
	rm -f $(BIN_NAME)
