DEFAULT_VERSION=0.1.0-local
VERSION := $(or $(VERSION),$(DEFAULT_VERSION))

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

all: build
build:
	$(GOBUILD) -ldflags "-w -X main.VERSION=$(VERSION)" -o './build/server' cmd/service/main.go
clean:
	rm -rf build


