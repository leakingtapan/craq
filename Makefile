# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build

bin:
	mkdir ./bin

build: bin
	$(GOBUILD) -o ./bin/craq ./cmd/craq/

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run: build
	./bin/craq

deps:
	$(GOGET) -v ./...
	$(GOMOD) tidy

lint:
	golangci-lint run


.PHONY: all build test clean run deps lint
