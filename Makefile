# Chimera Makefile
<<<<<<< HEAD
.PHONY: help build test clean install deps lint fmt vet integration-test docker

# Variables
BINARY_NAME=chimera
MAIN_PATH=./cmd
BUILD_DIR=./bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Build
build: deps ## Build the binary
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

# Install
install: build ## Install the binary
	go install $(LDFLAGS) $(MAIN_PATH)

# Development
fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, skipping lint check"; \
		echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Testing
test: ## Run unit tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

integration-test: ## Run integration tests
	@echo "Running Chimera integration tests..."
	@if [ -f "scripts/test-integration.sh" ]; then \
		chmod +x scripts/test-integration.sh; \
		scripts/test-integration.sh; \
	else \
		echo "⚠️  Integration test script not found, creating minimal test..."; \
		$(MAKE) build; \
		./bin/chimera --help >/dev/null && echo "✅ Basic CLI test passed" || echo "❌ Basic CLI test failed"; \
	fi

# Setup
setup: ## Setup development environment
	@echo "Setting up Chimera development environment..."
	@if [ -f ".chimera-codespaces" ]; then \
		echo "📦 Detected Codespaces environment"; \
	fi
	@# Initialize Go module if it doesn't exist
	@if [ ! -f "go.mod" ]; then \
		echo "Initializing Go module..."; \
		go mod init github.com/BigChiefRick/chimera; \
	fi
	$(MAKE) deps
	@echo "Installing development tools..."
	@# Install golangci-lint if not present
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@# Create necessary directories
	@mkdir -p bin pkg/discovery pkg/generation cmd/discover cmd/generate scripts examples test
	@# Make scripts executable
	@chmod +x scripts/*.sh 2>/dev/null || true
	@echo "✅ Development environment setup completed!"
	@# Run integration test if available
	@if [ -f "scripts/test-integration.sh" ]; then \
		echo "Running integration test..."; \
		$(MAKE) integration-test; \
	else \
		echo "💡 Tip: Run 'make build' to test your setup"; \
	fi

# Steampipe operations
steampipe-start: ## Start Steampipe service
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe service start; \
	else \
		echo "❌ Steampipe not found. Install with:"; \
		echo "   curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh | sudo sh"; \
	fi

steampipe-stop: ## Stop Steampipe service
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe service stop; \
	else \
		echo "❌ Steampipe not found"; \
	fi

steampipe-status: ## Check Steampipe service status
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe service status; \
	else \
		echo "❌ Steampipe not found"; \
	fi

steampipe-install-plugins: ## Install common Steampipe plugins
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe plugin install aws azure gcp kubernetes; \
	else \
		echo "❌ Steampipe not found"; \
	fi

steampipe-test: steampipe-start ## Test Steampipe with a simple query
	@if command -v steampipe >/dev/null 2>&1; then \
		echo "Testing Steampipe with simple query..."; \
		steampipe query "select 'Hello from Steampipe!' as message, current_timestamp as time"; \
	else \
		echo "❌ Steampipe not found"; \
	fi

# Terraformer operations
terraformer-test-aws: ## Test Terraformer with AWS
	@if command -v terraformer >/dev/null 2>&1; then \
		mkdir -p test-output/aws; \
		cd test-output/aws && terraform init; \
		terraformer import aws --resources=vpc --regions=us-east-1 --path-output=test-output/aws; \
	else \
		echo "❌ Terraformer not found. Install from: https://github.com/GoogleCloudPlatform/terraformer"; \
	fi

terraformer-test-azure: ## Test Terraformer with Azure
	@if command -v terraformer >/dev/null 2>&1; then \
		mkdir -p test-output/azure; \
		cd test-output/azure && terraform init; \
		terraformer import azure --resources=virtual_network --path-output=test-output/azure; \
	else \
		echo "❌ Terraformer not found"; \
	fi

terraformer-test-gcp: ## Test Terraformer with GCP
	@if command -v terraformer >/dev/null 2>&1; then \
		mkdir -p test-output/gcp; \
		cd test-output/gcp && terraform init; \
		terraformer import google --resources=networks --path-output=test-output/gcp; \
	else \
		echo "❌ Terraformer not found"; \
	fi

# Discovery tests
test-discovery-aws: steampipe-start ## Test discovery on AWS
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe query "select name, vpc_id, region from aws_vpc limit 5"; \
	else \
		echo "❌ Steampipe not found"; \
	fi
	
test-discovery-azure: steampipe-start ## Test discovery on Azure
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe query "select name, location from azure_resource_group limit 5"; \
	else \
		echo "❌ Steampipe not found"; \
	fi

test-discovery-gcp: steampipe-start ## Test discovery on GCP  
	@if command -v steampipe >/dev/null 2>&1; then \
		steampipe query "select name, project_id from gcp_project limit 5"; \
	else \
		echo "❌ Steampipe not found"; \
	fi

test-discovery-all: steampipe-start ## Test discovery on all clouds
	@if command -v steampipe >/dev/null 2>&1; then \
		echo "Testing multi-cloud discovery..."; \
		steampipe query "select 'AWS' as provider, count(*) as vpcs from aws_vpc union all select 'Azure' as provider, count(*) from azure_resource_group union all select 'GCP' as provider, count(*) from gcp_project"; \
	else \
		echo "❌ Steampipe not found"; \
	fi

# Codespaces helpers
codespaces-start: ## Start Codespaces development environment
	@if [ -f ".chimera-codespaces" ] || [ -n "$$CODESPACES" ]; then \
		if [ -f "scripts/codespaces.sh" ]; then \
			chmod +x scripts/codespaces.sh; \
			scripts/codespaces.sh start; \
		else \
			echo "🚀 Starting Chimera in Codespaces..."; \
			$(MAKE) build; \
			$(MAKE) steampipe-start; \
			echo "✅ Chimera is ready!"; \
		fi \
	else \
		echo "❌ Not in Codespaces environment"; \
		echo "💡 Run 'make steampipe-start' instead"; \
	fi

codespaces-demo: ## Run Codespaces demo
	@if [ -f ".chimera-codespaces" ] || [ -n "$$CODESPACES" ]; then \
		if [ -f "scripts/codespaces.sh" ]; then \
			chmod +x scripts/codespaces.sh; \
			scripts/codespaces.sh demo; \
		else \
			echo "🎬 Running Chimera demo..."; \
			$(MAKE) build; \
			./bin/chimera --help; \
			$(MAKE) steampipe-test; \
		fi \
	else \
		echo "❌ Not in Codespaces environment"; \
		$(MAKE) test; \
	fi

codespaces-status: ## Show Codespaces status
	@if [ -f "scripts/codespaces.sh" ]; then \
		chmod +x scripts/codespaces.sh; \
		scripts/codespaces.sh status; \
	else \
		echo "📊 Basic Project Status:"; \
		echo "  Location: $$(pwd)"; \
		echo "  Git branch: $$(git branch --show-current 2>/dev/null || echo 'unknown')"; \
		echo "  Binary exists: $$([ -f 'bin/chimera' ] && echo '✅ Yes' || echo '❌ No')"; \
		echo "  Go module: $$([ -f 'go.mod' ] && echo '✅ Yes' || echo '❌ No')"; \
	fi

# Development workflow
dev-build: ## Quick development build
	go build -o bin/chimera ./cmd

dev-test: ## Quick development test
	$(MAKE) dev-build
	./bin/chimera --help

dev-run: ## Build and run with help
	$(MAKE) dev-build
	./bin/chimera

# Tool installation helpers
install-tools: ## Install all development tools
	@echo "Installing development tools..."
	@# Install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "✅ golangci-lint installed"; \
	fi
	@# Install Steampipe
	@if ! command -v steampipe >/dev/null 2>&1; then \
		echo "Installing Steampipe..."; \
		curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh | sudo sh; \
		echo "✅ Steampipe installed"; \
	fi
	@# Install Terraformer
	@if ! command -v terraformer >/dev/null 2>&1; then \
		echo "Installing Terraformer..."; \
		TERRAFORMER_VERSION=$$(curl -s https://api.github.com/repos/GoogleCloudPlatform/terraformer/releases/latest | grep tag_name | cut -d '"' -f 4); \
		curl -L "https://github.com/GoogleCloudPlatform/terraformer/releases/download/$$TERRAFORMER_VERSION/terraformer-linux-amd64" -o terraformer; \
		sudo mv terraformer /usr/local/bin/terraformer; \
		sudo chmod +x /usr/local/bin/terraformer; \
		echo "✅ Terraformer installed"; \
	fi

check-tools: ## Check if development tools are installed
	@echo "Checking development tools..."
	@echo "Go: $$(command -v go >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "golangci-lint: $$(command -v golangci-lint >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "Steampipe: $$(command -v steampipe >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "Terraformer: $$(command -v terraformer >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "AWS CLI: $$(command -v aws >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "Azure CLI: $$(command -v az >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"
	@echo "gcloud: $$(command -v gcloud >/dev/null 2>&1 && echo '✅ Installed' || echo '❌ Missing')"

# Clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf test-output/
	rm -f coverage.out coverage.html
	go clean -cache

clean-all: clean ## Clean everything including Go module cache
	go clean -modcache

# Docker operations
docker-build: ## Build Docker image
	docker build -t chimera:$(VERSION) .

docker-run: ## Run Docker container
	docker run --rm -it chimera:$(VERSION)

# Quick start for new developers
quickstart: ## Quick start for new developers
	@echo "🚀 Chimera Quick Start"
	@echo "====================="
	$(MAKE) setup
	$(MAKE) build
	@echo ""
	@echo "✅ Setup complete! Try these commands:"
	@echo "  make help           # Show all available commands"
	@echo "  ./bin/chimera --help # Show CLI help"
	@echo "  make steampipe-test  # Test Steampipe integration"
	@echo "  make check-tools     # Check installed tools"

# Show project status
status: ## Show project status
	@echo "📊 Chimera Project Status"
	@echo "========================"
	@echo "Version: $(VERSION)"
	@echo "Binary: $$([ -f 'bin/chimera' ] && echo '✅ Built' || echo '❌ Not built')"
	@echo "Tests: $$(make test >/dev/null 2>&1 && echo '✅ Passing' || echo '❌ Failing')"
	@echo "Git status: $$(git status --porcelain | wc -l) modified files"
	@$(MAKE) check-tools
=======
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

# Quick development targets
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
>>>>>>> f199d23 (✅ Fix Codespace setup - Go environment working)
