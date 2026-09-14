.PHONY: build test vet fmt fmt-check check

build:
	go build ./cmd/hosta

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	test -z "$$(gofmt -l .)"

check: fmt-check vet test build
