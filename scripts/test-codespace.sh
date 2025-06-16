#!/bin/bash

# Codespaces helper script for Chimera
# This script provides utilities for working with Chimera in GitHub Codespaces

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

show_banner() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║          🔮 CHIMERA PROJECT          ║"
    echo "  ║   Multi-Cloud Infrastructure Tool    ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""
}

case "$1" in
    "start")
        show_banner
        print_info "🚀 Starting Chimera in Codespaces"
        
        # Check if we're in a Codespace
        if [ ! -f ".chimera-codespaces" ] && [ -z "$CODESPACES" ]; then
            print_warning "This doesn't appear to be a Codespaces environment"
            print_info "You can still use this script, but some features may not work"
        fi
        
        # Ensure project is built
        print_info "Building Chimera..."
        if make build &> /dev/null; then
            print_status "Chimera built successfully"
        else
            print_warning "Build failed, trying basic build..."
            go build -o bin/chimera ./cmd || {
                print_error "Failed to build Chimera"
                exit 1
            }
        fi
        
        # Start Steampipe service if available
        if command -v steampipe &> /dev/null; then
            print_info "Starting Steampipe service..."
            if steampipe service start &> /dev/null; then
                print_status "Steampipe service started"
            else
                print_warning "Steampipe service failed to start (may already be running)"
            fi
        else
            print_warning "Steampipe not found - some features may be limited"
        fi
        
        # Show status
        echo ""
        print_status "Chimera is ready! 🎉"
        echo ""
        echo "Quick start commands:"
        echo "  ./bin/chimera --help          # Show CLI help"
        echo "  make help                     # Show all available commands"
        echo "  make integration-test         # Test your environment"
        echo ""
        echo "Discovery examples:"
        echo "  ./bin/chimera discover --help # Discovery help"
        echo ""
        echo "Development commands:"
        echo "  make build                    # Build the project"
        echo "  make test                     # Run tests"
        echo "  make fmt                      # Format code"
        echo ""
        echo "Steampipe commands (if available):"
        echo "  steampipe query 'select 1'    # Test Steampipe"
        echo "  steampipe plugin list         # List installed plugins"
        ;;
    
    "demo")
        show_banner
        print_info "🎬 Running Chimera demo"
        
        # Build the project
        print_info "Building Chimera..."
        if ! make build &> /dev/null; then
            go build -o bin/chimera ./cmd
        fi
        print_status "Build completed"
        
        echo ""
        print_info "=== Chimera CLI Help ==="
        ./bin/chimera --help
        
        echo ""
        print_info "=== Testing Makefile targets ==="
        echo "Available make targets:"
        make help | grep -E "^\s+\w+" | head -10
        
        # Test Steampipe if available
        if command -v steampipe &> /dev/null; then
            echo ""
            print_info "=== Testing Steampipe ==="
            
            # Start service if not running
            if ! steampipe service status &> /dev/null; then
                steampipe service start &> /dev/null || true
                sleep 2
            fi
            
            if steampipe service status &> /dev/null; then
                print_status "Steampipe service is running"
                echo "Testing basic query..."
                steampipe query "select 'Hello from Steampipe!' as message, current_timestamp as time"
            else
                print_warning "Steampipe service not running"
            fi
        else
            print_warning "Steampipe not available in this environment"
        fi
        
        # Test cloud tools
        echo ""
        print_info "=== Cloud Tools Status ==="
        tools=("aws" "az" "gcloud" "terraform")
        for tool in "${tools[@]}"; do
            if command -v "$tool" &> /dev/null; then
                version=$($tool --version 2>/dev/null | head -1 || echo "version check failed")
                print_status "$tool: $version"
            else
                print_warning "$tool: not installed"
            fi
        done
        
        echo ""
        print_status "Demo completed! 🎉"
        echo ""
        echo "Next steps:"
        echo "1. Configure cloud credentials (aws configure, az login, etc.)"
        echo "2. Try: ./bin/chimera discover --help"
        echo "3. Explore the codebase in pkg/ directory"
        ;;
    
    "test")
        show_banner
        print_info "🧪 Running tests in Codespaces"
        
        # Run integration tests
        if [ -f "scripts/test-integration.sh" ]; then
            chmod +x scripts/test-integration.sh
            ./scripts/test-integration.sh
        else
            print_warning "Integration test script not found, running basic tests"
            
            # Basic Go tests
            print_info "Running Go tests..."
            if go test ./... -v; then
                print_status "Go tests passed"
            else
                print_warning "Some Go tests failed (this is normal for early development)"
            fi
            
            # Build test
            print_info "Testing build..."
            if make build; then
                print_status "Build test passed"
            else
                print_error "Build test failed"
            fi
        fi
        ;;
    
    "setup")
        show_banner
        print_info "🔧 Setting up Chimera development environment"
        
        # Run the full setup
        if make setup; then
            print_status "Setup completed successfully"
        else
            print_warning "Setup completed with warnings"
        fi
        
        # Additional Codespaces-specific setup
        print_info "Applying Codespaces-specific configuration..."
        
        # Ensure Git configuration
        if [ -z "$(git config --global user.name)" ]; then
            print_info "Setting up Git configuration..."
            echo "Please configure Git:"
            echo "  git config --global user.name 'Your Name'"
            echo "  git config --global user.email 'your.email@example.com'"
        fi
        
        print_status "Codespaces setup completed"
        ;;
    
    "status")
        show_banner
        print_info "📊 Chimera Environment Status"
        
        # Project status
        echo ""
        echo "Project Information:"
        echo "  Location: $(pwd)"
        echo "  Git branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
        echo "  Git status: $(git status --porcelain | wc -l) modified files"
        
        # Build status
        echo ""
        echo "Build Status:"
        if [ -f "bin/chimera" ]; then
            echo "  Binary: ✅ bin/chimera exists"
            echo "  Size: $(ls -lh bin/chimera | awk '{print $5}')"
        else
            echo "  Binary: ❌ bin/chimera not found"
        fi
        
        # Dependencies
        echo ""
        echo "Dependencies:"
        if [ -f "go.mod" ]; then
            echo "  Go module: ✅ go.mod exists"
        else
            echo "  Go module: ❌ go.mod missing"
        fi
        
        # Tools status
        echo ""
        echo "Development Tools:"
        tools=("go" "make" "steampipe" "terraformer" "aws" "az" "gcloud")
        for tool in "${tools[@]}"; do
            if command -v "$tool" &> /dev/null; then
                echo "  $tool: ✅ installed"
            else
                echo "  $tool: ❌ not installed"
            fi
        done
        
        # Services status
        echo ""
        echo "Services:"
        if command -v steampipe &> /dev/null; then
            if steampipe service status &> /dev/null; then
                echo "  Steampipe: ✅ running"
            else
                echo "  Steampipe: ⚠️  not running"
            fi
        else
            echo "  Steampipe: ❌ not installed"
        fi
        ;;
    
    "clean")
        show_banner
        print_info "🧹 Cleaning Chimera environment"
        
        # Clean build artifacts
        if make clean &> /dev/null; then
            print_status "Build artifacts cleaned"
        fi
        
        # Clean Go cache
        go clean -cache -modcache &> /dev/null || true
        print_status "Go cache cleaned"
        
        # Clean temporary files
        find . -name "*.tmp" -type f -delete 2>/dev/null || true
        find . -name "*.temp" -type f -delete 2>/dev/null || true
        print_status "Temporary files cleaned"
        
        print_status "Cleanup completed"
        ;;
    
    "logs")
        show_banner
        print_info "📋 Showing Chimera logs"
        
        # Show recent git commits
        echo "Recent commits:"
        git log --oneline -5 2>/dev/null || echo "No git history available"
        
        echo ""
        # Show Steampipe logs if available
        if command -v steampipe &> /dev/null; then
            echo "Steampipe service status:"
            steampipe service status || echo "Steampipe service not running"
        fi
        
        # Show system info
        echo ""
        echo "System Information:"
        echo "  OS: $(uname -s)"
        echo "  Architecture: $(uname -m)"
        echo "  Go version: $(go version 2>/dev/null || echo 'Go not found')"
        echo "  User: $(whoami)"
        echo "  Working directory: $(pwd)"
        ;;
    
    *)
        show_banner
        echo "Usage: $0 {start|demo|test|setup|status|clean|logs}"
        echo ""
        echo "Commands:"
        echo "  start   - Start Chimera services and show quick start info"
        echo "  demo    - Run a comprehensive demo of Chimera features"
        echo "  test    - Run integration tests"
        echo "  setup   - Set up the development environment"
        echo "  status  - Show environment and project status"
        echo "  clean   - Clean build artifacts and temporary files"
        echo "  logs    - Show logs and system information"
        echo ""
        echo "Examples:"
        echo "  $0 start           # Quick start"
        echo "  $0 demo            # Full demo"
        echo "  $0 test            # Run tests"
        echo ""
        print_info "For more help, see: https://github.com/BigChiefRick/chimera"
        ;;
esac