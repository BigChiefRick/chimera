# Chimera Phase 2 - Multi-Cloud Quick Start Guide

**🎉 PHASE 2 IMPLEMENTATION COMPLETE** - Real multi-cloud infrastructure discovery across AWS, Azure, and GCP!

## 🎯 What's New in Phase 2

### ✅ Multi-Cloud Provider Support
- **AWS Discovery** - Enhanced from Phase 1 with improved error handling
- **Azure Discovery** - Complete Azure Resource Manager integration
- **GCP Discovery** - Full Google Cloud Platform support
- **Cross-Cloud Operations** - Unified discovery across multiple providers

### ✅ Enhanced Features
- **Multi-Provider CLI** - Single command to discover across all clouds
- **Provider-Specific Options** - Dedicated flags for each cloud platform
- **Credential Management** - Seamless integration with cloud CLI tools
- **Unified Output** - Consistent resource format across all providers

## 🚀 Quick Start Examples

### 1. Single Cloud Discovery

```bash
# AWS Discovery (enhanced from Phase 1)
./bin/chimera discover --provider aws --region us-east-1 --format table

# Azure Discovery (NEW in Phase 2)
./bin/chimera discover --provider azure --azure-subscription "12345678-1234-1234-1234-123456789012" --region eastus --format table

# GCP Discovery (NEW in Phase 2)  
./bin/chimera discover --provider gcp --gcp-project "my-gcp-project" --region us-central1 --format table
```

### 2. Multi-Cloud Discovery (NEW!)

```bash
# Discover across AWS + Azure
./bin/chimera discover \
  --provider aws \
  --provider azure --azure-subscription "12345678-1234-1234-1234-123456789012" \
  --region us-east-1 --region eastus \
  --format table

# Discover across all three clouds
./bin/chimera discover \
  --provider aws \
  --provider azure --azure-subscription "12345678-1234-1234-1234-123456789012" \
  --provider gcp --gcp-project "my-gcp-project" \
  --region us-east-1 --region eastus --region us-central1 \
  --format json
```

### 3. Multi-Cloud Output Example

```bash
$ ./bin/chimera discover --provider aws --provider azure --azure-subscription "demo-sub" --format table

🔍 Multi-Cloud Infrastructure Discovery (Phase 2)
================================================

🔍 Discovering AWS resources...
✅ Found 6 AWS resources

🔍 Discovering AZURE resources...
✅ Found 4 Azure resources

🎉 Multi-Cloud Discovery Complete!
Total resources found: 10
Discovery duration: 2.1s

📊 Resource Summary by Provider:
   AWS: 6 resources
   AZURE: 4 resources

PROVIDER   NAME                     TYPE                          ID                       REGION          ZONE           
AWS        Hub-VPC                  aws_vpc                       vpc-03bc078b8ebc41abc    us-east-1                      
AWS        Production-Subnet        aws_subnet                    subnet-04682dfa9d873eb0f us-east-1       us-east-1b     
AWS        WebServer                aws_instance                  i-039b48c9fe902739c      us-east-1       us-east-1b     
AZURE      Production-RG            azure_resource_group          /subscriptions/.../rg     eastus                         
AZURE      Hub-VNet                 azure_virtual_network         /subscriptions/.../vnet   eastus                         
AZURE      WebApp-VM                azure_virtual_machine         /subscriptions/.../vm     eastus                         

Total: 10 resources
```

## 🔧 Prerequisites

### Cloud CLI Tools
You'll need the appropriate CLI tools for each cloud you want to discover:

```bash
# AWS CLI
aws --version
aws configure  # or aws configure sso

# Azure CLI  
az --version
az login

# Google Cloud CLI
gcloud --version
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### Build Requirements
- **Go 1.21+** - For building the application
- **Make** - For build automation

## 🏗️ Installation

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Install Phase 2 dependencies and build
make setup
make build

# Verify Phase 2 installation
make phase2-test
```

### 2. Test Cloud Credentials

```bash
# Test all cloud credentials
make test-all-creds

# Test individual clouds
make aws-test-creds
make azure-test-creds  
make gcp-test-creds
```

### 3. Run Your First Multi-Cloud Discovery

```bash
# Automatic multi-cloud discovery (uses available credentials)
make multi-cloud-discover

# Or run manual discovery
./bin/chimera discover --provider aws --region us-east-1 --format table
```

## 🎯 Resource Types Supported

### AWS Resources
- **VPCs** - Virtual Private Clouds
- **Subnets** - VPC subnets  
- **Security Groups** - Network security rules
- **EC2 Instances** - Virtual machines

### Azure Resources (NEW)
- **Resource Groups** - Resource containers
- **Virtual Networks** - VNets and address spaces
- **Subnets** - VNet subnets
- **Network Security Groups** - NSGs and rules
- **Virtual Machines** - Azure VMs

### GCP Resources (NEW)
- **Networks** - VPC networks
- **Subnetworks** - VPC subnets
- **Firewalls** - Firewall rules
- **Instances** - Compute Engine VMs

## 🔐 Credential Management

### AWS Credentials
```bash
# Option 1: AWS CLI profiles
aws configure
export AWS_PROFILE=my-profile

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_DEFAULT_REGION="us-east-1"

# Option 3: AWS SSO
aws configure sso
```

### Azure Credentials
```bash
# Interactive login (recommended)
az login

# Service principal (for automation)
az login --service-principal --username APP_ID --password PASSWORD --tenant TENANT_ID

# Get your subscription ID
az account show --query id --output tsv
```

### GCP Credentials
```bash
# Interactive login
gcloud auth login

# Service account (for automation)  
gcloud auth activate-service-account --key-file=path/to/service-account.json

# Set default project
gcloud config set project YOUR_PROJECT_ID
```

## 🎛️ Advanced Usage

### Resource Type Filtering

```bash
# Discover only VPCs/VNets across clouds
./bin/chimera discover \
  --provider aws \
  --provider azure --azure-subscription "sub-id" \
  --resource-type vpc \
  --resource-type virtual_network

# Discover only compute instances
./bin/chimera discover \
  --provider aws \
  --provider gcp --gcp-project "project-id" \
  --resource-type instance
```

### Region-Specific Discovery

```bash
# Multiple regions per provider
./bin/chimera discover \
  --provider aws --region us-east-1 --region us-west-2 \
  --provider azure --azure-subscription "sub-id" --region eastus --region westus2 \
  --format table
```

### Output Formats

```bash
# JSON output with full metadata
./bin/chimera discover --provider aws --format json

# Table output for human reading
./bin/chimera discover --provider aws --format table

# Save to file
./bin/chimera discover --provider aws --output multi-cloud-resources.json
```

## 🛠️ Development Commands

### Quick Testing
```bash
# Build and test Phase 2
make quickstart

# Run comprehensive demo
make demo

# Test specific cloud provider
make aws-discover-real
make azure-discover-real  
make gcp-discover-real
```

### Integration Testing
```bash
# Run all Phase 2 tests
make phase2-test

# Test multi-cloud integration
make integration-test

# Performance testing
make perf-test
```

## 🐛 Troubleshooting

### Common Issues

#### "Azure subscription ID is required"
```bash
# Solution: Provide subscription ID
az account show --query id --output tsv  # Get your subscription ID
./bin/chimera discover --provider azure --azure-subscription "YOUR_SUB_ID"
```

#### "GCP project ID is required"
```bash
# Solution: Set project and provide project ID
gcloud config set project YOUR_PROJECT_ID
./bin/chimera discover --provider gcp --gcp-project "YOUR_PROJECT_ID"
```

#### "No resources found"
This is normal if your account/regions have no resources. Try:
```bash
# Different regions
./bin/chimera discover --provider aws --region us-west-2

# Dry run to see what would be scanned
./bin/chimera discover --provider aws --dry-run
```

#### Authentication Issues
```bash
# Verify cloud credentials
make test-all-creds

# Re-authenticate if needed
aws configure
az login  
gcloud auth login
```

## 📊 Performance Expectations

### Discovery Times
- **Single provider** (small environment): 1-3 seconds
- **Multi-provider** (medium environment): 3-8 seconds  
- **Large enterprise** (100+ resources): 10-30 seconds

### Resource Scaling
- **Concurrent API calls**: Up to 10 per provider (configurable)
- **Memory usage**: ~100MB for 1000 resources across all clouds
- **Network efficiency**: Optimized API calls per resource type

## 🎉 What's Next

### Phase 3 Preview
Phase 2 lays the foundation for Phase 3 features:
- **IaC Generation** - Convert discovered resources to Terraform/Pulumi
- **VMware vSphere** - On-premises virtualization discovery
- **KVM/Libvirt** - Linux virtualization support
- **Resource Dependencies** - Understand resource relationships

### Contributing to Phase 3
Ready to contribute? The multi-cloud architecture makes adding new features straightforward:

```bash
# Check what's planned
cat README.md | grep -A 10 "Phase 3"

# Development workflow
make dev-build
make test
git checkout -b feature/your-feature
```

## 📞 Getting Help

- **📚 Documentation**: [README.md](README.md)
- **🐛 Issues**: [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/BigChiefRick/chimera/discussions)

---

**🎊 Congratulations!** You now have full multi-cloud infrastructure discovery working across AWS, Azure, and GCP.

**Ready for Phase 3?** Let's add IaC generation to complete the infrastructure reverse-engineering pipeline!

```bash
# Start exploring Phase 3
make status
make demo-real
```

*Built with ❤️ for cloud engineers managing multi-cloud environments*
