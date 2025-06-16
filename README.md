# Chimera

**Multi-Cloud Infrastructure Discovery and IaC Generation Tool**

A production-ready tool that connects to multiple cloud environments, discovers infrastructure resources, and generates Infrastructure as Code templates from existing infrastructure.

[![Phase 2 Complete](https://img.shields.io/badge/Phase%202-Complete-brightgreen.svg)](https://github.com/BigChiefRick/chimera)
[![Multi-Cloud Discovery](https://img.shields.io/badge/AWS%20%7C%20Azure%20%7C%20GCP-Working-blue.svg)](https://github.com/BigChiefRick/chimera)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 🎯 What Chimera Does

- **🔍 Reverse Engineer Infrastructure** - Convert existing cloud resources into manageable IaC
- **☁️ Multi-Cloud Support** - Work across AWS, Azure, GCP, VMware vSphere, and KVM environments  
- **📋 Standardize Management** - Generate consistent IaC templates across different platforms
- **⚡ Accelerate Migration** - Quickly codify existing infrastructure for modernization efforts

## 🚀 Phase 2 - Multi-Cloud Discovery Complete

**✅ PHASE 2 COMPLETE** - Full multi-cloud discovery across AWS, Azure, and GCP

### What's Working Now:

- **🔍 AWS Discovery** - VPCs, Subnets, Security Groups, EC2 Instances
- **🔍 Azure Discovery** - Resource Groups, Virtual Networks, Subnets, NSGs, Virtual Machines
- **🔍 GCP Discovery** - Networks, Subnetworks, Firewalls, Compute Instances
- **🖥️ Professional CLI** - Multi-cloud command structure with provider-specific flags
- **🏗️ Unified Architecture** - Consistent resource format across all cloud providers
- **📊 Multiple Output Formats** - JSON, YAML, Table formats
- **⚙️ Configuration Management** - YAML-based config with validation
- **🔐 Multi-Cloud Authentication** - Native integration with AWS, Azure, and GCP CLIs
- **📈 Performance Monitoring** - Execution timing and resource counting
- **🧪 Comprehensive Testing** - Integration tests and validation

### Real Multi-Cloud Discovery Example:

```bash
# Discover across AWS, Azure, and GCP in a single command
./bin/chimera discover \
  --provider aws \
  --provider azure --azure-subscription "12345678-1234-1234-1234-123456789012" \
  --provider gcp --gcp-project "my-gcp-project" \
  --region us-east-1 --region eastus --region us-central1 \
  --format table

# Output:
🔍 Multi-Cloud Infrastructure Discovery (Phase 2)
================================================

🔍 Discovering AWS resources...
✅ Found 6 AWS resources

🔍 Discovering AZURE resources...
✅ Found 4 Azure resources

🔍 Discovering GCP resources...
✅ Found 3 GCP resources

🎉 Multi-Cloud Discovery Complete!
Total resources found: 13
Discovery duration: 3.2s

📊 Resource Summary by Provider:
   AWS: 6 resources
   AZURE: 4 resources
   GCP: 3 resources

PROVIDER   NAME                TYPE                    REGION          
AWS        Hub-VPC            aws_vpc                 us-east-1       
AZURE      Production-RG      azure_resource_group    eastus          
GCP        default           gcp_compute_network     us-central1     

Total: 13 resources
```

## 🏗️ Architecture

```
┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐
│   Cloud Providers   │    │       Chimera        │    │   IaC Templates     │
│                     │───▶│   Discovery &       │───▶│    Generated        │
│ AWS ✅ Azure ✅    │    │   Generation Engine  │    │                     |
│ GCP ✅ VMware ⏳   │    │                      │    │ Terraform ⏳        |
│ KVM ⏳             │    │                      │    │ Pulumi ⏳           │
└─────────────────────┘    └───────────────────── ┘    └─────────────────────┘
```

### Core Components:

- **Discovery Engine** (`pkg/discovery/`) - Multi-provider resource discovery orchestration
- **AWS Provider** (`pkg/discovery/providers/aws.go`) - Production AWS connector
- **Azure Provider** (`pkg/discovery/providers/azure.go`) - Complete Azure ARM integration
- **GCP Provider** (`pkg/discovery/providers/gcp.go`) - Full Google Cloud support
- **Generation Framework** (`pkg/generation/`) - IaC template generation (Phase 3)
- **CLI Interface** (`cmd/`) - Professional command-line interface
- **Configuration** (`pkg/config/`) - YAML-based configuration management

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **Cloud CLIs configured** - AWS CLI, Azure CLI, Google Cloud CLI
- **Git** - For cloning the repository

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Build the project
make build

# Verify installation
./bin/chimera --help
```

### 2. Configure Cloud Access

Choose the cloud providers you want to discover:

#### AWS Setup
```bash
aws configure
# OR for SSO
aws configure sso
```

#### Azure Setup
```bash
az login
# Get your subscription ID
az account show --query id --output tsv
```

#### GCP Setup
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### 3. Test Your Setup

```bash
# Verify cloud credentials
make test-all-creds

# Test multi-cloud discovery
./bin/chimera discover --provider aws --provider azure --azure-subscription "your-sub-id" --format table
```

## 📖 Usage

### Multi-Cloud Discovery Commands

```bash
# Discover all resources across multiple clouds
./bin/chimera discover \
  --provider aws \
  --provider azure --azure-subscription "12345678-1234-1234-1234-123456789012" \
  --provider gcp --gcp-project "my-project" \
  --format table

# Single cloud discovery
./bin/chimera discover --provider aws --region us-east-1 --format table
./bin/chimera discover --provider azure --azure-subscription "sub-id" --region eastus --format table
./bin/chimera discover --provider gcp --gcp-project "project-id" --region us-central1 --format table

# Specific resource types across clouds
./bin/chimera discover \
  --provider aws --provider azure --azure-subscription "sub-id" \
  --resource-type vpc --resource-type virtual_network

# Multiple regions per provider
./bin/chimera discover \
  --provider aws --region us-east-1 --region us-west-2 \
  --provider gcp --gcp-project "project-id" --region us-central1 --region europe-west1

# Save multi-cloud results
./bin/chimera discover \
  --provider aws --provider azure --azure-subscription "sub-id" \
  --output multi-cloud-infrastructure.json

# Dry run (show plan without executing)
./bin/chimera discover \
  --provider aws --provider azure --azure-subscription "sub-id" \
  --dry-run
```

### Configuration Commands

```bash
# Initialize configuration file
./bin/chimera config init

# Validate current configuration
./bin/chimera config validate

# Show current configuration
./bin/chimera config show
```

### Output Formats

- **JSON** - Detailed machine-readable format with full metadata
- **Table** - Human-readable format for quick inspection
- **YAML** - Coming in Phase 3

## 🔍 Supported Resources by Provider

### AWS Resources
- **VPCs** - Virtual Private Clouds with CIDR blocks, state, default status
- **Subnets** - Including availability zones, IP counts, public IP mapping
- **Security Groups** - With ingress/egress rule counts and descriptions
- **EC2 Instances** - Including instance types, states, IPs, launch times

### Azure Resources
- **Resource Groups** - Resource containers with provisioning state
- **Virtual Networks** - VNets with address spaces and subnets
- **Subnets** - VNet subnets with address prefixes
- **Network Security Groups** - NSGs with security rule counts
- **Virtual Machines** - Azure VMs with size, state, and image information

### GCP Resources
- **Networks** - VPC networks with routing configuration
- **Subnetworks** - VPC subnets with CIDR ranges and regions
- **Firewalls** - Firewall rules with direction and priority
- **Compute Instances** - VM instances with machine types and network configs

## 🛠️ Development

### Build and Test

```bash
# Clean build
make clean && make build

# Run tests
make test

# Run Phase 2 integration tests
make phase2-test

# Test multi-cloud discovery
make multi-cloud-discover

# Format code
make fmt

# Lint code
make lint
```

### Development with GitHub Codespaces

This project is optimized for GitHub Codespaces with pre-configured multi-cloud development environment:

1. Open in Codespaces from GitHub
2. Run `make setup` to initialize
3. Configure cloud credentials (aws configure, az login, gcloud auth login)
4. Start developing!

### Development Helpers

```bash
# Quick development cycle
make dev-build && ./bin/chimera --help

# Run comprehensive demo
make demo

# Test all cloud credentials
make test-all-creds

# Check project status
make status
```

## 📊 Performance

Chimera is designed for performance and scalability across multiple clouds:

- **Fast Discovery** - 2-5 seconds for typical multi-cloud environments
- **Concurrent Processing** - Parallel discovery across providers
- **Memory Efficient** - Streams results without loading everything into memory
- **Error Resilient** - Continues discovery even if some providers fail
- **Provider Isolation** - One cloud failure doesn't stop others

## 🗂️ Project Structure

```
chimera/
├── cmd/                    # CLI commands and main entry point
│   ├── main.go            # Main CLI application
│   ├── discover/          # Multi-cloud discovery command
│   └── generate/          # Generation command (Phase 3)
├── pkg/                   # Core libraries
│   ├── discovery/         # Discovery engine and providers
│   │   ├── engine.go      # Multi-provider orchestration
│   │   ├── interfaces.go  # Core discovery interfaces
│   │   └── providers/     # Cloud provider implementations
│   │       ├── aws.go     # AWS discovery connector
│   │       ├── azure.go   # Azure discovery connector
│   │       └── gcp.go     # GCP discovery connector
│   ├── generation/        # IaC generation framework
│   │   └── interfaces.go  # Generation interfaces (Phase 3)
│   └── config/           # Configuration management
├── scripts/              # Build and development scripts
├── examples/             # Usage examples and demos
└── docs/                 # Documentation
```

## 🔧 Configuration

Chimera uses YAML configuration files. Create `~/.chimera.yaml`:

```yaml
# Global settings
debug: false
verbose: false
output_format: "json"
timeout: "10m"

# Discovery settings
discovery:
  max_concurrency: 10

# Provider configurations
providers:
  aws:
    regions: ["us-east-1", "us-west-2"]
    # Credentials read from AWS CLI/environment
  
  azure:
    subscription_id: "12345678-1234-1234-1234-123456789012"
    locations: ["eastus", "westus2"]
    # Credentials read from Azure CLI/environment
    
  gcp:
    project_id: "my-gcp-project"
    regions: ["us-central1", "us-east1"]
    # Credentials read from gcloud CLI/environment
```

Initialize with: `./bin/chimera config init`

## 🚧 Roadmap

### ✅ Phase 1: Foundation (COMPLETE)
- [x] Multi-cloud architecture design
- [x] AWS discovery connector with real resource scanning
- [x] Professional CLI framework
- [x] Configuration management system
- [x] Comprehensive testing and validation

### ✅ Phase 2: Multi-Cloud Discovery (COMPLETE)
- [x] Azure connector with Azure Resource Manager integration
- [x] GCP connector with Compute Engine integration
- [x] Multi-provider CLI orchestration
- [x] Cross-cloud resource aggregation
- [x] Provider-specific authentication and configuration

### 🔄 Phase 3: IaC Generation (Next)
- [ ] Terraform template generation from discovered resources
- [ ] Pulumi program generation
- [ ] CloudFormation template generation
- [ ] Cross-cloud module organization and dependencies
- [ ] Resource relationship mapping

### 🔄 Phase 4: Advanced Platforms
- [ ] VMware vSphere connector
- [ ] KVM/libvirt connector
- [ ] Kubernetes resource discovery
- [ ] Resource diffing and change detection
- [ ] State management integration

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes and test: `make test && make phase2-test`
4. Commit: `git commit -am "Add your feature"`
5. Push: `git push origin feature/your-feature`
6. Create a Pull Request

### Running Tests

```bash
# Unit tests
make test

# Multi-cloud integration tests (requires cloud credentials)
make phase2-test

# Test specific cloud providers
make aws-discover-real
make azure-discover-real
make gcp-discover-real

# Test all cloud credentials
make test-all-creds
```

## 📄 License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.

## 🔒 Security

Please report security vulnerabilities through GitHub Security Advisories or email the maintainers directly.

## 🙏 Acknowledgments

This project builds upon excellent open-source tools:

- **[Terraformer](https://github.com/GoogleCloudPlatform/terraformer)** - Infrastructure reverse engineering inspiration
- **[Steampipe](https://steampipe.io)** - Multi-cloud SQL interface concept
- **[AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)** - Robust AWS integration
- **[Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)** - Comprehensive Azure support
- **[Google Cloud SDK](https://cloud.google.com/sdk)** - Complete GCP integration
- **[Cobra CLI](https://cobra.dev/)** - Professional CLI framework
- **[Viper](https://github.com/spf13/viper)** - Configuration management

## 📞 Support

- **📚 Documentation**: [GitHub Wiki](https://github.com/BigChiefRick/chimera/wiki)
- **🐛 Issues**: [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/BigChiefRick/chimera/discussions)

---

**🎉 Phase 2 Complete!** Ready to discover and codify your multi-cloud infrastructure.

```bash
# Get started with multi-cloud discovery now!
git clone https://github.com/BigChiefRick/chimera.git
cd chimera && make build

# Configure your cloud credentials
aws configure && az login && gcloud auth login

# Discover across all your clouds
./bin/chimera discover --provider aws --provider azure --azure-subscription "your-sub-id" --provider gcp --gcp-project "your-project" --format table
```

*Built with ❤️ for cloud engineers managing multi-cloud environments*
