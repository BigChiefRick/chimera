#!/bin/bash

# Chimera Codespaces Setup Script
# This script sets up the Chimera development environment in GitHub Codespaces

set -e

echo "🔧 Setting up Chimera development environment in Codespaces..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Create marker file to indicate this is set up for Codespaces
touch .chimera-codespaces

print_info "Installing Steampipe..."
# Install Steampipe
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"

print_info "Installing Terraformer..."
# Install Terraformer
TERRAFORMER_VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/terraformer/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -LO "https://github.com/GoogleCloudPlatform/terraformer/releases/download/${TERRAFORMER_VERSION}/terraformer-linux-amd64"
chmod +x terraformer-linux-amd64
sudo mv terraformer-linux-amd64 /usr/local/bin/terraformer

print_info "Setting up Go environment..."
# Ensure Go is properly configured
echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~/.bashrc
echo 'export GOPATH=/go' >> ~/.bashrc
echo 'export GO111MODULE=on' >> ~/.bashrc

print_info "Installing Go development tools..."
# Install Go tools
go install -a github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install -a github.com/cosmtrek/air@latest  # Live reload for Go
go install -a golang.org/x/tools/gopls@latest # Go language server

print_info "Setting up project structure..."
# Initialize the project if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    go mod init github.com/BigChiefRick/chimera
    print_status "Go module initialized"
fi

# Create necessary directories
mkdir -p {cmd/{discover,generate},pkg/{discovery/{steampipe,providers},generation/{terraformer,templates},config,credentials},configs,scripts,docs,examples,test,bin}

# Create placeholder files
touch cmd/discover/.gitkeep cmd/generate/.gitkeep pkg/discovery/providers/.gitkeep pkg/generation/templates/.gitkeep configs/.gitkeep docs/.gitkeep examples/.gitkeep test/.gitkeep bin/.gitkeep

print_info "Installing Steampipe plugins..."
# Install common Steampipe plugins
steampipe plugin install aws azure gcp kubernetes

print_info "Creating Codespaces-specific configuration..."
# Create a Codespaces-specific config
cat > .vscode/settings.json << 'EOF'
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.vetOnSave": "package",
  "go.buildOnSave": "off",
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "go.gocodeAutoBuild": false,
  "go.buildTags": "",
  "go.toolsGopath": "/go",
  "terminal.integrated.defaultProfile.linux": "bash",
  "files.associations": {
    "*.tf": "terraform",
    "*.tfvars": "terraform",
    "Makefile": "makefile"
  },
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
EOF

print_info "Setting up development helpers..."
# Create a Codespaces-specific development script
cat > scripts/codespaces.sh << 'EOF'
#!/bin/bash

# Codespaces-specific helper script for Chimera

case "$1" in
    "start")
        echo "🚀 Starting Chimera development environment..."
        make steampipe-start
        echo "✅ Environment ready!"
        echo "📝 Try: make test-discovery-aws"
        ;;
    "setup-aws")
        echo "Setting up AWS credentials for Codespaces..."
        echo "Please run: gh auth login --scopes codespace"
        echo "Then set your AWS credentials as Codespace secrets:"
        echo "  - AWS_ACCESS_KEY_ID"
        echo "  - AWS_SECRET_ACCESS_KEY"
        echo "  - AWS_DEFAULT_REGION"
        ;;
    "setup-azure")
        echo "Setting up Azure for Codespaces..."
        echo "Run: az login --use-device-code"
        ;;
    "setup-gcp")
        echo "Setting up GCP for Codespaces..."
        echo "Run: gcloud auth login"
        echo "Then: gcloud config set project YOUR_PROJECT_ID"
        ;;
    "test")
        echo "Running quick development test..."
        make fmt
        make vet
        echo "✅ Code looks good!"
        ;;
    "demo")
        echo "Running Chimera demo..."
        echo "1. Testing Steampipe connection..."
        steampipe query "select 'Chimera is ready!' as message" || echo "❌ Steampipe not ready"
        
        echo "2. Testing Terraformer..."
        terraformer version || echo "❌ Terraformer not ready"
        
        echo "3. Testing Go build..."
        make build || echo "❌ Build failed"
        
        echo "✅ Demo complete!"
        ;;
    "help"|*)
        echo "Chimera Codespaces Helper"
        echo "========================"
        echo ""
        echo "Usage: ./scripts/codespaces.sh <command>"
        echo ""
        echo "Commands:"
        echo "  start      - Start development environment"
        echo "  setup-aws  - Guide for AWS setup"
        echo "  setup-azure- Guide for Azure setup"
        echo "  setup-gcp  - Guide for GCP setup"
        echo "  test       - Quick code quality check"
        echo "  demo       - Run environment demo"
        echo "  help       - Show this help"
        ;;
esac
EOF

chmod +x scripts/codespaces.sh

print_info "Creating environment README..."
# Create a Codespaces-specific README
cat > .devcontainer/README.md << 'EOF'
# Chimera Codespaces Environment

This Codespace is pre-configured with everything you need to develop Chimera:

## 🛠️ Pre-installed Tools

- **Go 1.21+** - Development language
- **Terraform** - Infrastructure as Code
- **Steampipe** - Multi-cloud SQL interface
- **Terraformer** - Reverse engineering tool
- **AWS CLI** - Amazon Web Services
- **Azure CLI** - Microsoft Azure
- **Google Cloud CLI** - Google Cloud Platform
- **Docker** - Containerization

## 🚀 Quick Start

1. **Initialize the project:**
   ```bash
   make setup
   ```

2. **Start Steampipe:**
   ```bash
   ./scripts/codespaces.sh start
   ```

3. **Configure cloud credentials** (choose your providers):
   ```bash
   # AWS
   ./scripts/codespaces.sh setup-aws
   
   # Azure
   ./scripts/codespaces.sh setup-azure
   
   # GCP
   ./scripts/codespaces.sh setup-gcp
   ```

4. **Test everything:**
   ```bash
   ./scripts/codespaces.sh demo
   ```

## 🔧 Development Workflow

```bash
# Build the project
make build

# Run tests
make test

# Format code
make fmt

# Run integration tests
make integration-test

# Quick development helper
./scripts/codespaces.sh test
```

## 🌐 Port Forwarding

The following ports are automatically forwarded:
- **9193** - Steampipe PostgreSQL interface
- **8080** - Chimera API (when implemented)
- **3000** - Development server (when needed)

## 📁 Persistent Storage

These directories are persisted across Codespace rebuilds:
- `/go/pkg/mod` - Go module cache
- `~/.steampipe` - Steampipe configuration and cache

## 🔐 Cloud Credentials

### AWS
Set these as Codespace secrets:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_DEFAULT_REGION`

### Azure
Use device code authentication:
```bash
az login --use-device-code
```

### Google Cloud
Use browser authentication:
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

## 🆘 Troubleshooting

**Steampipe connection issues:**
```bash
steampipe service restart
steampipe query "select 'test' as message"
```

**Go module issues:**
```bash
go mod tidy
go mod download
```

**Tool not found:**
```bash
# Restart the Codespace or run:
.devcontainer/setup.sh
```
EOF

print_info "Setting up Git configuration..."
# Configure Git for Codespaces
git config --global init.defaultBranch main
git config --global pull.rebase false

print_info "Downloading Go dependencies..."
go mod tidy || print_warning "Go mod tidy failed - will retry after first build"

print_status "Codespaces environment setup complete! 🎉"

echo ""
echo "🌟 Welcome to Chimera development in Codespaces!"
echo ""
echo "Quick start commands:"
echo "  make setup                    # Full development setup"
echo "  ./scripts/codespaces.sh start # Start development environment"
echo "  ./scripts/codespaces.sh demo  # Test everything"
echo ""
echo "📖 See .devcontainer/README.md for detailed instructions"
