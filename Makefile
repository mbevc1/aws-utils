.PHONY: help
.DEFAULT: help
ifndef VERBOSE
.SILENT:
endif

NO_COLOR=\033[0m
GREEN=\033[32;01m
YELLOW=\033[33;01m
RED=\033[31;01m

VER?=dev
#GHASH:=$(shell git rev-parse --short HEAD)
#VERSION?=$(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)
GOTELEMETRY:=	off
GO:=            go
#GO_BUILD:=      go build -mod vendor -ldflags "-s -w -X main.GitCommit=${GHASH} -X main.Version=${VERSION}"
GO_BUILD:=      go build -mod mod -ldflags "-s -w -X main.GitCommit=${GHASH} -X main.Version=${VERSION}"
#VERSION="${VERSION}" goreleaser --snapshot --rm-dist
GO_VENDOR:=     go mod vendor
BIN:=           aws-utils

help:: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[33m%-20s\033[0m %s\n", $$1, $$2}'

#$(BIN): vendor ## Produce binary
$(BIN): ## Produce binary
	GO111MODULE=on $(GO_BUILD)

# We always want vendor to run
.PHONY: vendor test
vendor: **/*.go ## Build vendor deps
	GO111MODULE=on $(GO_VENDOR)

clean: clean-vendor ## Clean artefacts
	rm -rf $(BIN) $(BIN)_* $(BIN).exe dist/

clean-vendor: ## Clean vendor folder
	rm -rf vendor

clean-cache: ## Clean Golang mod cache
	go clean --modcache
	go clean --cache

build: clean $(BIN) ## Build binary
	upx $(BIN)

run:
	go run .

snapshot: clean ## Build local snapshot
	#goreleaser build --snapshot --clean
	goreleaser build --clean --snapshot --single-target

dev: clean ## Dev test target
	go build -ldflags "-s -w -X main.Version=${VER}" -o $(BIN)
	upx $(BIN)

test: ## Run tests
	go test -v ./...

fmt: **/*.go ## Formt Golang code
	go fmt ./...

lint:
	golint ./...

vet:
	go vet -all ./...

$(BIN)_linux_amd64: vendor **/*.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $@ *.go
	upx $@

$(BIN)_linux_alpine: vendor **/*.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ *.go
	upx $@

$(BIN)_darwin_amd64: vendor **/*.go
	GOOS=darwin go build -o $@ *.go
	upx $@

$(BIN)_windows_amd64.exe: vendor **/*.go
	GOOS=windows GOARCH=amd64 go build -o $@ *.go
	upx $@

pack: $(BIN)_linux_amd64 $(BIN)_darwin_amd64 $(BIN)_windows_amd64.exe
	zip $(BIN)_linux_amd64.zip $(BIN)_linux_amd64
	zip $(BIN)_darwin_amd64.zip $(BIN)_darwin_amd64
	zip $(BIN)_windows_amd64.zip $(BIN)_windows_amd64.exe

fmtcheck: vendor **/*.go ## Check formatting
	@gofmt_files=$$(gofmt -l `find . -name '*.go' | grep -v vendor`); \
	    if [ -n "$${gofmt_files}" ]; then \
	    	echo 'gofmt needs running on the following files:'; \
	    	echo "$${gofmt_files}"; \
	    	echo "You can use the command: \`make fmt\` to reformat code."; \
	    	exit 1; \
	    fi; \
	    exit 0
