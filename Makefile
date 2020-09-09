BINDIR      := $(CURDIR)/bin
BINNAME     ?= hiver

GOPATH        = $(shell go env GOPATH)
GOIMPORTS     = $(GOPATH)/bin/goimports
ARCH          = $(shell uname -p)

SRC        := $(shell find . -type f -name '*.go' -print)

# Required for globs to work correctly
SHELL      = /usr/bin/env bash

.PHONY: all
all: build

.PHONY: build
build:
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BINDIR)/$(BINNAME) ./cmd/$(BINNAME)

.PHONY: install
install:
	GO111MODULE=on CGO_ENABLED=0 go install ./cmd/$(BINNAME)

.PHONY: clean
clean:
	rm $(BINDIR)/$(BINNAME)

