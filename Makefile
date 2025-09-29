.PHONY: help build run test lint fmt mocks docker-build docker-run clean tidy tools verify

APP_NAME := mcp-server-time
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

help: ## Show available commands
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o $(APP_NAME) ./cmd/main.go

run: ## Run the application locally
	go run ./cmd/main.go

test: ## Run all tests
	@echo ">>> Running tests"
	@go test ./...

test-unit: ## Run unit tests only
	@echo ">>> Running unit tests"
	@go test -short ./...

test-integration: ## Run integration tests only
	@echo ">>> Running integration tests"
	@go test -run Integration ./...

lint: ## Run linters
	@echo ">>> Linting (go vet)"
	@go vet ./...

fmt: ## Format code
	@echo ">>> Formatting (gofmt)"
	@gofmt -s -w .

tidy: ## Tidy go modules
	@echo ">>> Tidying go modules"
	@go mod tidy

tools: ## Install development tools
	@echo ">>> Installing development tools"
	@go install go.uber.org/mock/mockgen@latest

mocks: ## Generate mocks
	@echo ">>> Generating mocks"
	@go generate ./...

verify: fmt lint test build ## Run all verification steps

docker-build: ## Build Docker image
	docker buildx build --platform linux/amd64,linux/arm64 -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-build-local: ## Build Docker image locally
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 -p 9090:9090 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		$(APP_NAME):latest

clean: ## Clean build artifacts
	rm -f $(APP_NAME)
	docker rmi $(APP_NAME):$(VERSION) $(APP_NAME):latest 2>/dev/null || true

dev: ## Run with file watching (requires air)
	air

install-dev-tools: ## Install development tools
	go install github.com/air-verse/air@latest
	go install go.uber.org/mock/mockgen@latest
