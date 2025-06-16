#!/bin/bash

# Chimera Project Initialization Script
# This script initializes the Chimera project structure and dependencies

set -e

echo "🚀 Initializing Chimera Project"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print status
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

# Check if we're in the right directory
if [ ! -f ".chimera-root" ]; then
    print_error "Please run this script from the root of the Chimera repository"
    exit 1
fi

print_info "Initializing Go module..."
if [ ! -f "go.mod" ]; then
    go mod init github.com/BigChiefRick/chimera
    print_status "Go module initialized"
else
    print_status "Go module already exists"
fi

print_info "Creating project directory structure..."

# Create directory structure
directories=(
    "cmd/discover"
    "cmd/generate"
    "pkg/discovery/steampipe"
    "pkg/discovery/providers"
    "pkg/generation/terraformer"
    "pkg/generation/templates"
    "pkg/config"
    "pkg/credentials"
    "configs"
    "scripts"
    "docs"
    "examples"
    "test"
    "bin"
)

for dir in "${directories[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        print_status "Created directory: $dir"
    else
        print_status "Directory exists: $dir"
    fi
done

print_info "Creating placeholder files..."

# Create placeholder files to maintain directory structure
placeholder_files=(
    "cmd/discover/.gitkeep"
    "cmd/generate/.gitkeep"
    "pkg/discovery/providers/.gitkeep"
    "pkg/generation/templates/.gitkeep"
    "configs/.gitkeep"
    "docs/.gitkeep"
    "examples/.gitkeep"
    "test/.gitkeep"
    "bin/.gitkeep"
)

for file in "${placeholder_files[@]}"; do
    if [ ! -f "$file" ]; then
        touch "$file"
        print_status "Created placeholder: $file"
    fi
done

print_info "Setting up .gitignore..."
if [ ! -f ".gitignore" ]; then
    cat > .gitignore << 'EOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
!bin/.gitkeep

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.html

# Go workspace file
go.work

# Dependency directories
vendor/

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Temporary files
*.tmp
*.temp
.cache/

# Logs
*.log
logs/

# Configuration files with secrets
config.yaml
config.json
*.env
.env
.env.local

# Generated IaC files during testing
generated/
test-output/
*.tf
*.tf.json
terraform.tfstate*
.terraform/

# Steampipe files
.steampipe/

# Chimera specific
chimera-test*/
chimera-output*/
EOF
    print_status "Created .gitignore"
else
    print_status ".gitignore already exists"
fi

print_info "Setting up basic configuration files..."

# Create example configuration
if [ ! -f "configs/example.yaml" ]; then
    cat > configs/example.yaml << 'EOF'
# Chimera Configuration Example
# Copy this file to ~/.chimera.yaml and customize

# Global settings
debug: false
verbose: false
output_format: "json"
timeout: "10m"

# Discovery settings
discovery:
  max_concurrency: 10
  steampipe:
    host: "localhost"
    port: 9193
    database: "steampipe"
    user: "steampipe"
    timeout: "30s"

# Generation settings
generation:
  output_path: "./generated"
  organize_by_type: true
  include_state: true
  validate_output: true

# Provider configurations
providers:
  aws:
    regions: ["us-east-1", "us-west-2"]
    # Credentials will be read from AWS CLI/environment
  
  azure:
    # Credentials will be read from Azure CLI/environment
    
  gcp:
    # Credentials will be read from gcloud CLI/environment
    
  vmware:
    # Connection details for vSphere
    
  kvm:
    # Connection details for KVM hosts
EOF
    print_status "Created example configuration"
fi

print_info "Downloading Go dependencies..."
go mod tidy
print_status "Dependencies downloaded and organized"

print_info "Making scripts executable..."
chmod +x scripts/*.sh
print_status "Scripts are now executable"

print_info "Creating development helpers..."

# Create a simple development script
if [ ! -f "scripts/dev.sh" ]; then
    cat > scripts/dev.sh << 'EOF'
#!/bin/bash

# Development helper script for Chimera

case "$1" in
    "build")
        echo "Building Chimera..."
        make build
        ;;
    "test")
        echo "Running tests..."
        make test
        ;;
    "integration")
        echo "Running integration tests..."
        make integration-test
        ;;
    "steampipe")
        echo "Starting Steampipe..."
        make steampipe-start
        ;;
    "fmt")
        echo "Formatting code..."
        make fmt
        ;;
    *)
        echo "Usage: $0 {build|test|integration|steampipe|fmt}"
        echo ""
        echo "Available commands:"
        echo "  build       - Build the Chimera binary"
        echo "  test        - Run unit tests"
        echo "  integration - Run integration tests"
        echo "  steampipe   - Start Steampipe service"
        echo "  fmt         - Format Go code"
        ;;
esac
EOF
    chmod +x scripts/dev.sh
    print_status "Created development helper script"
fi

echo ""
print_status "Chimera project initialization completed! 🎉"
echo ""
echo "Next steps:"
echo "1. Run: make setup                    # Install development tools"
echo "2. Run: make integration-test         # Test your environment"
echo "3. Run: scripts/dev.sh steampipe      # Start Steampipe service"
echo "4. Configure cloud credentials:"
echo "   - AWS: aws configure"
echo "   - Azure: az login"
echo "   - GCP: gcloud auth login"
echo "5. Start developing! Check out the examples/ directory"
echo ""
print_info "Documentation: https://github.com/BigChiefRick/chimera"
print_info "Happy coding! 🚀"