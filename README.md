# Chimera

**Multi-Cloud Infrastructure Discovery and IaC Generation Tool**

A production-ready tool that connects to multiple cloud and virtualization environments, discovers infrastructure resources, and generates Infrastructure as Code templates from existing infrastructure.

[![Phase 1 Complete](https://img.shields.io/badge/Phase%201-Complete-brightgreen.svg)](https://github.com/BigChiefRick/chimera)
[![AWS Discovery](https://img.shields.io/badge/AWS-Working-blue.svg)](https://github.com/BigChiefRick/chimera)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 🎯 What Chimera Does

- **🔍 Reverse Engineer Infrastructure** - Convert existing cloud resources into manageable IaC
- **☁️ Multi-Cloud Support** - Work across AWS, Azure, GCP, VMware vSphere, and KVM environments  
- **📋 Standardize Management** - Generate consistent IaC templates across different platforms
- **⚡ Accelerate Migration** - Quickly codify existing infrastructure for modernization efforts

## 🚀 Phase 1 - Production Ready

**✅ PHASE 1 COMPLETE** - Fully functional AWS discovery with real infrastructure scanning

### What's Working Now:

- **🔍 AWS Discovery** - VPCs, Subnets, Security Groups, EC2 Instances
- **🖥️ Professional CLI** - Full command structure with help, configuration, validation
- **🏗️ Multi-Cloud Architecture** - Extensible framework ready for Azure, GCP, VMware, KVM
- **📊 Multiple Output Formats** - JSON, YAML, Table formats
- **⚙️ Configuration Management** - YAML-based config with validation
- **🔐 AWS SSO Support** - Works with AWS profiles and temporary credentials
- **📈 Performance Monitoring** - Execution timing and resource counting
- **🧪 Comprehensive Testing** - Integration tests and validation

### Real Discovery Example:

```bash
# Discover all AWS resources in us-east-2
./bin/chimera discover --provider aws --region us-east-2 --format table

# Output:
🔍 Discovering Real AWS Resources
================================
🔍 Target region: us-east-2
🔑 Validating AWS credentials...
✅ AWS credentials validated successfully!
🔍 Scanning for AWS resources...
🎉 Discovery Complete! Found 6 resources
   Duration: 1.38s

NAME                 TYPE                 ID                       REGION          ZONE           
Hub-VPC              aws_vpc              vpc-03bc078b8ebc41abc    us-east-2                      
Production-Subnet    aws_subnet           subnet-04682dfa9d873eb0f us-east-2       us-east-2b     
Dev-Subnet           aws_subnet           subnet-0be5db5318542785d us-east-2       us-east-2c     
Management-Subnet    aws_subnet           subnet-09dc1ff0092b0b585 us-east-2       us-east-2a     
default              aws_security_group   sg-0b2c57cfcd7f348dd     us-east-2                      
rsmith               aws_instance         i-039b48c9fe902739c      us-east-2       us-east-2b     

Total: 6 resources
```

## 🏗️ Architecture

```
┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐
│   Cloud Providers   │    │       Chimera        │    │   IaC Templates     │
│                     │───▶│   Discovery &        │───▶│    Generated        │
│ AWS ✅ Azure ⏳     │    │   Generation Engine   │    │                     │
│ GCP ⏳ VMware ⏳    │    │                      │    │ Terraform ⏳       │
│ KVM ⏳              │    │                      │    │ Pulumi ⏳          │
└─────────────────────┘    └──────────────────────┘    └─────────────────────┘
```

### Core Components:

- **Discovery Engine** (`pkg/discovery/`) - Multi-provider resource discovery orchestration
- **AWS Provider** (`pkg/discovery/providers/aws.go`) - Production AWS connector
- **Generation Framework** (`pkg/generation/`) - IaC template generation (Phase 2)
- **CLI Interface** (`cmd/`) - Professional command-line interface
- **Configuration** (`pkg/config/`) - YAML-based configuration management

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **AWS CLI configured** - `aws configure` or AWS SSO setup
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

### 2. Configure AWS Access

Choose one of these methods:

#### Option A: AWS CLI (Recommended)
```bash
aws configure
# OR for SSO
aws configure sso
```

#### Option B: Environment Variables
```bash
export AWS_PROFILE=your-profile-name
# OR
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

#### Option C: AWS SSO Profile
```bash
export AWS_PROFILE=AdministratorAccess-123456789012
```

### 3. Test Your Setup

```bash
# Verify AWS credentials
aws sts get-caller-identity

# Test Chimera discovery
./bin/chimera discover --provider aws --region us-east-1 --format table
```

## 📖 Usage

### Discovery Commands

```bash
# Discover all resources in a region
./bin/chimera discover --provider aws --region us-east-1

# Table format for easy reading
./bin/chimera discover --provider aws --region us-east-1 --format table

# Discover specific resource types
./bin/chimera discover --provider aws --region us-east-1 --resource-type vpc --resource-type instance

# Multiple regions
./bin/chimera discover --provider aws --region us-east-1 --region us-west-2

# Save to file
./bin/chimera discover --provider aws --region us-east-1 --output infrastructure.json

# Dry run (show plan without executing)
./bin/chimera discover --provider aws --region us-east-1 --dry-run
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
- **YAML** - Coming in Phase 2

### Supported AWS Resources

- **VPCs** - Virtual Private Clouds with CIDR blocks, state, default status
- **Subnets** - Including availability zones, IP counts, public IP mapping
- **Security Groups** - With ingress/egress rule counts and descriptions
- **EC2 Instances** - Including instance types, states, IPs, launch times

## 🛠️ Development

### Build and Test

```bash
# Clean build
make clean && make build

# Run tests
make test

# Run integration tests
make integration-test

# Verify Phase 1 completion
make phase1-test

# Format code
make fmt

# Lint code
make lint
```

### Development with GitHub Codespaces

This project is optimized for GitHub Codespaces with pre-configured development environment:

1. Open in Codespaces from GitHub
2. Run `make setup` to initialize
3. Configure AWS credentials
4. Start developing!

### Development Helpers

```bash
# Quick development cycle
make dev-build && ./bin/chimera --help

# Run demo
make demo

# Check project status
make status
```

## 📊 Performance

Chimera is designed for performance and scalability:

- **Fast Discovery** - 1-2 seconds for typical AWS regions
- **Concurrent Processing** - Configurable concurrent resource discovery
- **Memory Efficient** - Streams results without loading everything into memory
- **Error Resilient** - Continues discovery even if some resources fail

## 🗂️ Project Structure

```
chimera/
├── cmd/                    # CLI commands and main entry point
│   ├── main.go            # Main CLI application
│   ├── discover/          # Discovery command implementation
│   └── generate/          # Generation command (Phase 2)
├── pkg/                   # Core libraries
│   ├── discovery/         # Discovery engine and providers
│   │   ├── engine.go      # Multi-provider orchestration
│   │   ├── interfaces.go  # Core discovery interfaces
│   │   └── providers/     # Cloud provider implementations
│   │       └── aws.go     # AWS discovery connector
│   ├── generation/        # IaC generation framework
│   │   └── interfaces.go  # Generation interfaces (Phase 2)
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
```

Initialize with: `./bin/chimera config init`

## 🚧 Roadmap

### ✅ Phase 1: Foundation (COMPLETE)
- [x] Multi-cloud architecture design
- [x] AWS discovery connector with real resource scanning
- [x] Professional CLI framework
- [x] Configuration management system
- [x] Comprehensive testing and validation

### 🔄 Phase 2: Multi-Cloud (In Progress)
- [ ] Azure connector (`pkg/discovery/providers/azure.go`)
- [ ] GCP connector (`pkg/discovery/providers/gcp.go`)
- [ ] Enhanced multi-region discovery
- [ ] Resource relationship mapping

### 🔄 Phase 3: IaC Generation
- [ ] Terraform template generation
- [ ] Pulumi program generation  
- [ ] CloudFormation template generation
- [ ] Module organization and dependencies

### 🔄 Phase 4: Advanced Features
- [ ] VMware vSphere connector
- [ ] KVM/libvirt connector
- [ ] Resource diffing and change detection
- [ ] State management integration

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes and test: `make test`
4. Commit: `git commit -am "Add your feature"`
5. Push: `git push origin feature/your-feature`
6. Create a Pull Request

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires AWS credentials)
make integration-test

# Phase 1 completion verification
make phase1-test
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
- **[Cobra CLI](https://cobra.dev/)** - Professional CLI framework
- **[Viper](https://github.com/spf13/viper)** - Configuration management

## 📞 Support

- **📚 Documentation**: [GitHub Wiki](https://github.com/BigChiefRick/chimera/wiki)
- **🐛 Issues**: [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/BigChiefRick/chimera/discussions)

---

**🎉 Phase 1 Complete!** Ready to discover and codify your cloud infrastructure.

```bash
# Get started now!
git clone https://github.com/BigChiefRick/chimera.git
cd chimera && make build
./bin/chimera discover --provider aws --region us-east-1 --format table
```

*Built with ❤️ for cloud engineers and DevOps teams*
