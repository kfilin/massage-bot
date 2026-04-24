# Vera Massage Bot - Development Command Center
BINARY_NAME=bot
BIN_DIR=bin

.PHONY: all build test run clean lint vet cover docker-up help

all: build test

## 🛠 Build & Run
build:
	@echo "🏗 Building binary..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/bot

run:
	@echo "🚀 Running bot locally..."
	go run ./cmd/bot

## 🧪 Testing & Quality
test:
	@echo "🧪 Running unit tests..."
	go test ./... -v

cover:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "💡 Run 'go tool cover -html=coverage.out' to see details in browser"

lint:
	@echo "🔍 Running golangci-lint..."
	golangci-lint run

vet:
	@echo "🩺 Running go vet..."
	go vet ./...

## 🧹 Cleanup
clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out
	rm -rf logs/*.log

## 🐳 Docker
docker-up:
	@echo "🐳 Starting environment..."
	docker-compose up -d --build

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
