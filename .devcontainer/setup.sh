#!/bin/bash

set -e

echo "🚀 Setting up Chimera Phase 2 development environment in Codespaces"
echo "=================================================================="

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

# Install AWS CLI v2
print_info "Installing AWS CLI..."
if ! command -v aws &> /dev/null; then
    cd /tmp
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" -s
    unzip -q awscliv2.zip
    sudo ./aws/install
    rm -rf aws awscliv2.zip
    cd - > /dev/null
    print_status "AWS CLI installed"
else
    print_status "AWS CLI already installed"
fi

# Install Azure CLI
print_info "Installing Azure CLI..."
if ! command -v az &> /dev/null; then
    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash > /dev/null 2>&1
    print_status "Azure CLI installed"
else
    print_status "Azure CLI already installed"
fi

# Install Google Cloud CLI
print_info "Installing Google Cloud CLI..."
if ! command -v gcloud &> /dev/null; then
    # Download and install gcloud
    cd /tmp
    curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-458.0.1-linux-x86_64.tar.gz
    tar -xf google-cloud-cli-458.0.1-linux-x86_64.tar.gz
    sudo mv google-cloud-sdk /opt/
    sudo ln -sf /opt/google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud
    sudo ln -sf /opt/google-cloud-sdk/bin/gsutil /usr/local/bin/gsutil
    rm google-cloud-cli-458.0.1-linux-x86_64.tar.gz
    cd - > /dev/null
    print_status "Google Cloud CLI installed"
else
    print_status "Google Cloud CLI already installed"
fi

# Install Terraform
print_info "Installing Terraform..."
if ! command -v terraform &> /dev/null; then
    cd /tmp
    TERRAFORM_VERSION="1.6.6"
    curl -fsSL "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" -o terraform.zip
    unzip -q terraform.zip
    sudo mv terraform /usr/local/bin/
    rm terraform.zip
    cd - > /dev/null
    print_status "Terraform installed"
else
    print_status "Terraform already installed"
fi

# Install Steampipe
print_info "Installing Steampipe..."
if ! command -v steampipe &> /dev/null; then
    curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh | sudo sh > /dev/null 2>&1
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
    
    curl -L "https://github.com/GoogleCloudPlatform/terraformer/releases/download/${TERRAFORMER_VERSION}/terraformer-linux-amd64" -o terraformer -s
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
touch cmd/discover/.gitkeep 2>/dev/null || true
touch cmd/generate/.gitkeep 2>/dev/null || true
touch pkg/discovery/providers/.gitkeep 2>/dev/null || true
touch pkg/generation/terraformer/.gitkeep 2>/dev/null || true
touch examples/.gitkeep 2>/dev/null || true
touch test/.gitkeep 2>/dev/null || true
touch bin/.gitkeep 2>/dev/null || true

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

# Verify all tools
print_info "Verifying installed tools..."
echo ""
echo "Installed Cloud CLI Versions:"
echo "  AWS CLI: $(aws --version 2>/dev/null || echo 'Not installed')"
echo "  Azure CLI: $(az --version 2>/dev/null | head -1 || echo 'Not installed')"
echo "  Google Cloud CLI: $(gcloud --version 2>/dev/null | head -1 || echo 'Not installed')"
echo "  Terraform: $(terraform --version 2>/dev/null | head -1 || echo 'Not installed')"
echo "  Steampipe: $(steampipe --version 2>/dev/null || echo 'Not installed')"
echo "  Terraformer: $(terraformer version 2>/dev/null || echo 'Not installed')"

echo ""
print_status "Codespaces Phase 2 setup completed! 🎉"
echo ""
echo "Next steps:"
echo "1. Configure cloud credentials:"
echo "   - AWS: aws configure"
echo "   - Azure: az login"
echo "   - GCP: gcloud auth login"
echo "2. Test setup: make test-all-creds"
echo "3. Run discovery: make multi-cloud-discover"
echo "4. Build and test: make build && make phase2-test"
echo ""
print_info "Happy multi-cloud development! 🚀"
