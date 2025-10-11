build:
	go build -o main main.go

run:
	./main

clean:
	rm main

clean-config:
	rm /Users/jarrodservilla/.config/slacky/config.yaml
	
clean-all: clean clean-config

dev: clean build run

lint:
	go mod tidy
	go fmt ./...
	golangci-lint run ./...

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

test-update-golden:
	UPDATE_GOLDEN=true go test ./...