.PHONY: build
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

build:
	go build -o stockscraper-${GO_OS}-${GO_ARCH}
