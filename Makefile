.PHONY: build run clean clean-config clean-all dev lint test test-verbose test-short test-coverage test-update

# Version info for local builds
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/jcserv/slacky/internal/version.Version=$(VERSION)

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o slacky main.go

run:
	./slacky

clean:
	rm -f slacky

clean-config:
	rm /Users/jarrodservilla/.config/slacky/config.yaml
	
clean-all: clean clean-config

dev: clean build run

lint:
	go mod tidy
	go fmt .
	golangci-lint run . --fix

test:
	go test -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...

test-verbose:
	go test -v -race ./...

test-short:
	go test -short ./...

test-coverage:
	go test -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-update:
	UPDATE_GOLDEN=true go test ./...