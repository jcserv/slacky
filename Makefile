build:
	go build -o slacky main.go

run:
	./slacky

clean:
	rm slacky

clean-config:
	rm /Users/jarrodservilla/.config/slacky/config.yaml
	
clean-all: clean clean-config

dev: clean build run