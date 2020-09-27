BENCH_PKG := github.com/tikv/pd
TEST_PKGS := $(shell find . -iname "*_test.go" -exec dirname {} \; | \
                     sort -u | sed -e "s/^\./github.com\/lhy1024\/bench/")

PACKAGES := go list ./...
PACKAGE_DIRECTORIES := $(PACKAGES) | sed 's|$(BENCH_PKG)/||'
GO_CHECKER := awk '{ print } END { if (NR > 0) { exit 1 } }'

GO_TOOLS_BIN_PATH := $(shell pwd)/.tools/bin
PATH := $(GO_TOOLS_BIN_PATH):$(PATH)
SHELL := env PATH='$(PATH)' GOBIN='$(GO_TOOLS_BIN_PATH)' /bin/bash

default: build

install-go-tools: export GO111MODULE=on
install-go-tools:
	@mkdir -p $(GO_TOOLS_BIN_PATH)
	@which golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_TOOLS_BIN_PATH) v1.27.0
	@grep '_' tools.go | sed 's/"//g' | awk '{print $$2}' | xargs go install

build:
	GO111MODULE=on go build -o ./bin/bench -v

test: install-go-tools
	GO111MODULE=on go test $(TEST_PKGS)

check:install-go-tools static lint tidy

static: export GO111MODULE=on
static:
	gofmt -s -l -d $$($(PACKAGE_DIRECTORIES)) 2>&1 | $(GO_CHECKER)
	golangci-lint run $$($(PACKAGE_DIRECTORIES))

lint:
	@echo "linting"
	revive -formatter friendly -config revive.toml $$($(PACKAGES))

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum