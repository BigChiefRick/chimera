# Chimera

A multi-cloud infrastructure discovery and Infrastructure as Code (IaC) generation tool that connects to multiple cloud and virtualization environments to automatically generate IaC templates from existing infrastructure.

## Overview

Chimera connects to your existing cloud and virtualization environments, discovers infrastructure resources, ingests metadata via APIs, and generates Infrastructure as Code templates. This enables you to:

- **Reverse Engineer Infrastructure** - Convert existing cloud resources into manageable IaC
- **Multi-Cloud Support** - Work across AWS, Azure, GCP, VMware vSphere, and KVM environments
- **Standardize Management** - Generate consistent IaC templates across different platforms
- **Accelerate Migration** - Quickly codify existing infrastructure for modernization efforts

## Supported Platforms

- ☁️ **Amazon Web Services (AWS)**
- ☁️ **Microsoft Azure**
- ☁️ **Google Cloud Platform (GCP)**
- 🖥️ **VMware vSphere**
- 🖥️ **KVM/libvirt**

## Supported IaC Outputs

- Terraform (.tf)
- Pulumi
- AWS CloudFormation
- Azure ARM Templates
- *(More formats planned)*

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Cloud Provider │    │     Chimera      │    │  IaC Templates  │
│   Environments  │───▶│   Discovery &    │───▶│   Generated     │
│                 │    │   Generation     │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.21+ or Python 3.9+
- Valid credentials for target cloud/virtualization platforms
- Appropriate permissions for resource discovery in target environments

### Installation

```bash
# Clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Install dependencies (coming soon)
# go mod download
# OR
# pip install -r requirements.txt
```

### Quick Start

```bash
# Configure credentials (coming soon)
chimera configure

# Discover infrastructure
chimera discover --provider aws --region us-east-1

# Generate IaC templates
chimera generate --output terraform --target ./output/
```

## Configuration

Chimera supports multiple credential management approaches:
- Environment variables
- Cloud provider CLI profiles
- Configuration files
- Integration with HashiCorp Vault
- Cloud-native secret managers

## Project Status

🚧 **Early Development** - This project is in active development. APIs and functionality are subject to change.

### Current Phase - Phase 1: Foundation
- [x] Research and identify core open-source components
- [ ] Terraformer integration and analysis
- [ ] Steampipe integration for unified discovery
- [ ] Core architecture design based on learnings
- [ ] Basic multi-cloud connector framework

### Roadmap
- **Phase 1**: AWS + Terraform support
- **Phase 2**: Azure and GCP connectors
- **Phase 3**: VMware vSphere integration
- **Phase 4**: KVM/libvirt support
- **Phase 5**: Multi-IaC tool support

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Create a feature branch
git checkout -b feature/your-feature-name

# Make your changes and commit
git commit -am "Add your feature"

# Push and create a pull request
git push origin feature/your-feature-name
```

## License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.

## Security

Please report security vulnerabilities to [security@yourproject.com] or through GitHub Security Advisories.

## Acknowledgments

## Acknowledgments

This project builds upon the excellent work of the open-source community, including:
- **[Terraformer](https://github.com/GoogleCloudPlatform/terraformer)** - CLI tool for reverse engineering infrastructure to Terraform
- **[Steampipe](https://steampipe.io)** - Universal SQL interface for cloud APIs and services  
- **[Former2](https://github.com/iann0036/former2)** - AWS infrastructure to CloudFormation/Terraform converter
- **[Prowler](https://github.com/prowler-cloud/prowler)** - Multi-cloud security assessment and resource discovery
- **[Azure Export for Terraform](https://github.com/Azure/aztfexport)** - Microsoft's official Azure to Terraform tool

---

**Note**: Chimera is designed to discover and codify existing infrastructure. Always review generated IaC templates before applying them to production environment