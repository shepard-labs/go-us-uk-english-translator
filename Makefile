.PHONY: all build test test-coverage lint clean help

# Binaries
CLI_BINARY=bin/translate
API_BINARY=bin/apiserver

all: clean build test

build: ## Build both CLI and API server binaries
	@echo "Building CLI binary..."
	go build -o $(CLI_BINARY) ./cmd/translate
	@echo "Building API server binary..."
	go build -o $(API_BINARY) ./cmd/apiserver
	@echo "Build complete. Binaries are in the bin/ directory."

test: ## Run unit and integration tests
	@echo "Running tests..."
	go test ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race ./...

test-coverage: ## Run tests and generate coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

clean: ## Remove build artifacts and coverage reports
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html

run-cli: build ## Run the CLI tool (e.g., make run-cli ARGS="--help")
	./$(CLI_BINARY) $(ARGS)

run-api: build ## Run the API server
	./$(API_BINARY)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
