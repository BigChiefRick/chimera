# Chimera Makefile
.PHONY: help build test clean deps

BINARY_NAME=chimera
BUILD_DIR=./bin
VERSION=v0.1.0-dev

help: ## Show help
	@echo 'Available targets:'
	@echo '  build     - Build the binary'
	@echo '  test      - Run tests'
	@echo '  clean     - Clean build artifacts'
	@echo '  deps      - Download dependencies'

deps: ## Download dependencies
	go mod tidy

build: deps ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	go clean -cache

dev-build: ## Quick build
	go build -o bin/chimera ./cmd

dev-test: ## Quick test
	$(MAKE) dev-build
	./bin/chimera --help

setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod init github.com/BigChiefRick/chimera || true
	$(MAKE) deps
	@echo "✅ Setup complete!"

check-tools: ## Check development tools
	@echo "Go: $$(go version)"
	@echo "Make: $$(make --version | head -1)"
