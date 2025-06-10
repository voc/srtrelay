.PHONY: all srtrelay install download-go

GO_VERSION := 1.24.4
GO_URL := https://go.dev/dl/go$(GO_VERSION).linux-amd64.tar.gz
GO_DIR := $(PWD)/gobuild/go
GOCACHE := $(PWD)/.cache/golang-build
GOMODCACHE := $(PWD)/.cache/golang-mod
GOENV := GOROOT=$(GO_DIR) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) PATH=$(GO_DIR)/bin:$$PATH

all: srtrelay

download-go:
	@mkdir -p gobuild
	@curl -sSL $(GO_URL) | tar -C gobuild -xzf -

srtrelay: download-go
	@$(GOENV) $(GO_DIR)/bin/go build -o srtrelay

install: srtrelay
	@mkdir -p $(PWD)/debian/srtrelay/usr/bin
	@install -m 0755 srtrelay $(PWD)/debian/srtrelay/usr/bin
