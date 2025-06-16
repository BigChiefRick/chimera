# Chimera Quick Start Guide

Get up and running with Chimera in minutes!

## Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Cloud CLI tools** (optional but recommended):
  - AWS CLI - `aws configure`
  - Azure CLI - `az login`
  - Google Cloud CLI - `gcloud auth login`

## Quick Setup

### 1. Clone and Initialize

```bash
# Clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Initialize the project
chmod +x scripts/init-project.sh
./scripts/init-project.sh
```

### 2. Set Up Development Environment

```bash
# Install development tools and run integration tests
make setup
```

### 3. Install Required Tools

```bash
# Install Steampipe (unified cloud querying)
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"

# Install Steampipe plugins
steampipe plugin install aws azure gcp

# Install Terraformer (IaC generation)
# For macOS:
brew install terraformer

# For Linux:
# See: https://github.com/GoogleCloudPlatform/terraformer/releases
```

### 4. Configure Cloud Access

```bash
# AWS (choose one)
aws configure                    # Interactive setup
export AWS_PROFILE=myprofile     # Use existing profile

# Azure
az login

# Google Cloud
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### 5. Test Your Setup

```bash
# Run comprehensive integration test
make integration-test

# Test individual components
make steampipe-start
steampipe query "select 'Hello from Steampipe!' as message"

# Test Terraformer
terraformer version
```

## First Discovery

### Start Steampipe Service

```bash
make steampipe-start
```

### Query Your Infrastructure

```bash
# List AWS VPCs
steampipe query "select name, vpc_id, cidr_block from aws_vpc"

# List Azure Resource Groups
steampipe query "select name, location from azure_resource_group"

# Cross-cloud resource summary
steampipe query "
  select 'AWS' as provider, count(*) as resources from aws_vpc
  union all
  select 'Azure' as provider, count(*) as resources from azure_resource_group
  union all
  select 'GCP' as provider, count(*) as resources from gcp_project
"
```

### Generate IaC from Existing Resources

```bash
# Create test workspace
mkdir chimera-test && cd chimera-test

# Test Terraformer with AWS VPCs
terraformer import aws --resources=vpc --regions=us-east-1

# Check generated files
ls -la
cat aws/vpc/vpc.tf
```

## Development Workflow

### Build and Test

```bash
# Format code
make fmt

# Run linting
make lint

# Run tests
make test

# Build binary
make build

# The binary will be in ./bin/chimera
./bin/chimera --help
```

### Development Helper

```bash
# Use the development helper script
./scripts/dev.sh build        # Build the project
./scripts/dev.sh test         # Run tests
./scripts/dev.sh integration  # Run integration tests
./scripts/dev.sh steampipe    # Start Steampipe
./scripts/dev.sh fmt          # Format code
```

## Project Structure

```
chimera/
├── cmd/                    # CLI commands
├── pkg/
│   ├── discovery/         # Resource discovery logic
│   ├── generation/        # IaC generation logic
│   ├── config/           # Configuration management
│   └── credentials/      # Credential handling
├── configs/              # Configuration files
├── scripts/              # Utility scripts
├── docs/                 # Documentation
├── examples/             # Usage examples
└── test/                 # Test files
```

## Next Steps

1. **Explore the Architecture**: Check out `pkg/discovery/interfaces.go` and `pkg/generation/interfaces.go`
2. **Run Examples**: Look in the `examples/` directory
3. **Contribute**: See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
4. **Report Issues**: Use [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)

## Troubleshooting

### Common Issues

**Steampipe connection failed:**
```bash
# Check if service is running
steampipe service status

# Restart service
steampipe service restart
```

**Terraformer not found:**
```bash
# Install manually
curl -LO https://github.com/GoogleCloudPlatform/terraformer/releases/latest/download/terraformer-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64
chmod +x terraformer-*
sudo mv terraformer-* /usr/local/bin/terraformer
```

**AWS credentials not working:**
```bash
# Check credentials
aws sts get-caller-identity

# Check Steampipe can access AWS
steampipe query "select account_id from aws_caller_identity"
```

### Getting Help

- 📚 **Documentation**: [GitHub Wiki](https://github.com/BigChiefRick/chimera/wiki)
- 🐛 **Issues**: [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/BigChiefRick/chimera/discussions)

---

**Ready to start building? Run the integration test and start exploring!** 🚀

```bash
make integration-test && echo "🎉 You're ready to go!"
```