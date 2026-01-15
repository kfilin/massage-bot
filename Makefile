.PHONY: test build run

test:
	go test ./... -v

build:
	go build -o bin/bot ./cmd/bot

run:
	go run ./cmd/bot
