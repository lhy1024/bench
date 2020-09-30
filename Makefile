BENCH_PKG := github.com/lhy1024/bench
TEST_PKGS := $(shell find . -iname "*_test.go" -exec dirname {} \; | \
                     sort -u | sed -e "s/^\./github.com\/lhy1024\/bench/")
PACKAGES := go list ./...
PACKAGE_DIRECTORIES := $(PACKAGES) | sed 's|$(BENCH_PKG)/||'
GO_CHECKER := awk '{ print } END { if (NR > 0) { exit 1 } }'
GO_TOOLS_BIN_PATH := $(shell pwd)/.tools/bin
PATH := $(GO_TOOLS_BIN_PATH):$(PATH)
SHELL := env PATH='$(PATH)' GOBIN='$(GO_TOOLS_BIN_PATH)' /bin/bash
OVERALLS := overalls
BUILD_BIN_PATH := $(shell pwd)/bin

default: build

install-go-tools: export GO111MODULE=on
install-go-tools:
	@mkdir -p $(GO_TOOLS_BIN_PATH)
	@which golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_TOOLS_BIN_PATH) v1.27.0
	@grep '_' tools.go | sed 's/"//g' | awk '{print $$2}' | xargs go install

ci: build check test

build:
	@echo "build"
	GO111MODULE=on go build -o $(BUILD_BIN_PATH)/bench cmd/main.go

test: install-go-tools
	@echo "test"
	GO111MODULE=on go test $(TEST_PKGS)

check:install-go-tools static lint tidy

static: export GO111MODULE=on
static:
	@echo "static"
	gofmt -s -l -d $$($(PACKAGE_DIRECTORIES)) 2>&1 | $(GO_CHECKER)
	golangci-lint run $$($(PACKAGE_DIRECTORIES))

lint:
	@echo "linting"
	revive -formatter friendly -config revive.toml $$($(PACKAGES))

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum

travis_coverage: export GO111MODULE=on
travis_coverage:
ifeq ("$(TRAVIS_COVERAGE)", "1")
	CGO_ENABLED=1 $(OVERALLS) -concurrency=8 -project=github.com/lhy1024/bench -covermode=count -ignore='.git,vendor' -- -coverpkg=./...
else
	@echo "coverage only runs in travis."
endif
