build:
	go build -o slacky ./cmd/slacky

run:
	./slacky

clean:
	rm slacky

clean-config:
	rm /Users/jarrodservilla/.config/slacky/config.yaml
	
clean-all: clean clean-config

dev: clean build run

lint:
	go mod tidy
	go fmt ./...
	golangci-lint run ./...

test:
	go test ./...