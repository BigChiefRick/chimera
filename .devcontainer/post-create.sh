#!/bin/bash

set -e

echo "🚀 Setting up Chimera Multi-Cloud Development Environment"
echo "======================================================="

# Install Google Cloud CLI via APT (most reliable)
echo "Installing Google Cloud CLI..."
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add - 2>/dev/null
sudo apt-get update > /dev/null 2>&1
sudo apt-get install -y google-cloud-cli > /dev/null 2>&1

# Install Steampipe
echo "Installing Steampipe..."
curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh | sudo sh > /dev/null 2>&1

# Install Terraformer
echo "Installing Terraformer..."
TERRAFORMER_VERSION="0.8.30"
curl -L "https://github.com/GoogleCloudPlatform/terraformer/releases/download/${TERRAFORMER_VERSION}/terraformer-linux-amd64" -o /tmp/terraformer -s
sudo mv /tmp/terraformer /usr/local/bin/terraformer
sudo chmod +x /usr/local/bin/terraformer

# Install Go development tools
echo "Installing Go tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Initialize Go module if needed
if [ ! -f "go.mod" ]; then
    go mod init github.com/BigChiefRick/chimera
fi

# Download dependencies
go mod tidy

# Build the project
make build

echo "✅ Chimera development environment ready!"
echo ""
echo "Next steps:"
echo "1. Configure cloud credentials:"
echo "   aws configure"
echo "   az login" 
echo "   gcloud auth login"
echo "2. Test setup: make test-all-creds"
echo "3. Run discovery: make multi-cloud-discover"
