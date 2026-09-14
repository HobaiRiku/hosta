APP_NAME    := hosta
PKG_VERSION := github.com/HobaiRiku/hosta/internal/version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
-X $(PKG_VERSION).Version=$(VERSION) \
-X $(PKG_VERSION).Commit=$(COMMIT) \
-X $(PKG_VERSION).BuildDate=$(DATE)

BUILD_DIR := build/bin

.PHONY: build run tidy test vet fmt fmt-check check clean print-version

build:
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/hosta

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/hosta

tidy:
	go mod tidy

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	test -z "$$(gofmt -l .)"

check: fmt-check vet test build

clean:
	rm -rf $(BUILD_DIR)

print-version:
	@echo "version=$(VERSION) commit=$(COMMIT) date=$(DATE)"
