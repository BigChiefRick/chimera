et -e

echo "🚀 Setting up Chimera development environment in Codespaces"
echo "=========================================================="

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

# Mark this as a Codespaces environment
touch .chimera-codespaces
print_status "Marked as Codespaces environment"

# Initialize Go module if it doesn't exist
print_info "Setting up Go module..."
if [ ! -f "go.mod" ]; then
    go mod init github.com/BigChiefRick/chimera
    print_status "Go module initialized"
else
    print_status "Go module already exists"
fi

# Download dependencies
print_info "Downloading Go dependencies..."
go mod tidy
print_status "Dependencies downloaded"

# Install development tools
print_info "Installing development tools..."

# Install golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    print_status "golangci-lint installed"
else
    print_status "golangci-lint already installed"
fi

# Install Steampipe
print_info "Installing Steampipe..."
if ! command -v steampipe &> /dev/null; then
    curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh | sudo sh
    print_status "Steampipe installed"
else
    print_status "Steampipe already installed"
fi

# Install Steampipe plugins
print_info "Installing Steampipe plugins..."
steampipe plugin install aws azure gcp kubernetes 2>/dev/null || {
    print_warning "Some Steampipe plugins failed to install (this is normal in Codespaces)"
}

# Install Terraformer
print_info "Installing Terraformer..."
if ! command -v terraformer &> /dev/null; then
    TERRAFORMER_VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/terraformer/releases/latest | grep tag_name | cut -d '"' -f 4 | head -1)
    if [ -z "$TERRAFORMER_VERSION" ]; then
        TERRAFORMER_VERSION="0.8.24"  # Fallback version
    fi
    
    curl -L "https://github.com/GoogleCloudPlatform/terraformer/releases/download/${TERRAFORMER_VERSION}/terraformer-linux-amd64" -o terraformer
    sudo mv terraformer /usr/local/bin/terraformer
    sudo chmod +x /usr/local/bin/terraformer
    print_status "Terraformer installed (version: $TERRAFORMER_VERSION)"
else
    print_status "Terraformer already installed"
fi

# Create necessary directories
print_info "Creating project directories..."
directories=(
    "cmd/discover"
    "cmd/generate"
    "pkg/discovery/providers"
    "pkg/generation/terraformer"
    "pkg/config"
    "pkg/credentials"
    "scripts"
    "examples"
    "test"
    "bin"
)

for dir in "${directories[@]}"; do
    mkdir -p "$dir"
done
print_status "Project directories created"

# Make scripts executable
print_info "Making scripts executable..."
chmod +x scripts/*.sh 2>/dev/null || true
print_status "Scripts made executable"

# Create .gitkeep files for empty directories
touch cmd/discover/.gitkeep
touch cmd/generate/.gitkeep
touch pkg/discovery/providers/.gitkeep
touch pkg/generation/terraformer/.gitkeep
touch examples/.gitkeep
touch test/.gitkeep
touch bin/.gitkeep

# Build the project to test everything works
print_info "Building project..."
if go build -o bin/chimera ./cmd 2>/dev/null; then
    print_status "Project builds successfully"
else
    print_warning "Project build failed (this is expected if main.go is incomplete)"
fi

# Start Steampipe service
print_info "Starting Steampipe service..."
steampipe service start 2>/dev/null || {
    print_warning "Steampipe service failed to start (this is normal in some environments)"
}

echo ""
print_status "Codespaces setup completed! 🎉"
echo ""
echo "Next steps:"
echo "1. Run: make help                    # Show available commands"
echo "2. Run: make setup                   # Complete development setup"
echo "3. Run: make integration-test        # Test your environment"
echo "4. Configure cloud credentials:"
echo "   - AWS: aws configure"
echo "   - Azure: az login"
echo "   - GCP: gcloud auth login"
echo ""
print_info "Happy coding! 🚀"
