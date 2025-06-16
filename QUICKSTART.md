# Chimera Quick Start Guide - Phase 1 Final

Get up and running with **real AWS infrastructure discovery** in minutes!

## 🎯 What You'll Achieve

After following this guide, you'll have:
- ✅ **Working AWS discovery** scanning real cloud resources
- ✅ **Professional CLI** with full command structure
- ✅ **Multiple output formats** (JSON, table)
- ✅ **Ready for Phase 2** development

## Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **AWS CLI configured** - Or AWS SSO access
- **Git** - For cloning the repository

## 🚀 5-Minute Setup

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/BigChiefRick/chimera.git
cd chimera

# Build the project (includes dependency download)
make build

# Verify build
./bin/chimera --help
```

**Expected Output:**
```
Chimera - Multi-cloud infrastructure discovery and IaC generation tool
Usage: chimera [command]
Commands:
  discover    Discover infrastructure resources
  generate    Generate Infrastructure as Code  
  config      Manage Chimera configuration
  version     Show version information
```

### 2. Configure AWS Access

Choose the method that matches your AWS setup:

#### Option A: AWS CLI (Most Common)
```bash
aws configure
# Enter your AWS Access Key ID, Secret, and region
```

#### Option B: AWS SSO (Enterprise)
```bash
aws configure sso
# Follow the SSO setup prompts
```

#### Option C: AWS Profile (Existing Setup)
```bash
# If you already have AWS profiles configured
export AWS_PROFILE=your-profile-name
```

#### Option D: Environment Variables
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

### 3. Test AWS Connectivity

```bash
# Verify AWS credentials work
aws sts get-caller-identity

# Should show your AWS account info
{
    "UserId": "AIDACKCEVSQ6C2EXAMPLE",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/YourUsername"
}
```

### 4. Run Your First Discovery

```bash
# Discover AWS resources in table format (easy to read)
./bin/chimera discover --provider aws --region us-east-1 --format table
```

**Expected Output:**
```
🔍 Attempting Real AWS Discovery
================================
🔍 Target region: us-east-1
🔑 Validating AWS credentials...
✅ AWS credentials validated successfully!
🔍 Scanning for AWS resources...
🎉 Discovery Complete! Found X resources
   Duration: 1.23s

NAME                 TYPE                 ID                       REGION          ZONE           
MyVPC                aws_vpc              vpc-12345678             us-east-1                      
PublicSubnet         aws_subnet           subnet-abcdef12          us-east-1       us-east-1a     
WebServer            aws_instance         i-0123456789abcdef0      us-east-1       us-east-1a     

Total: X resources
```

## 🎉 Success! You're Now Running Real Discovery

### Try Different Output Formats

```bash
# JSON format (detailed, machine-readable)
./bin/chimera discover --provider aws --region us-east-1 --format json

# Save to file
./bin/chimera discover --provider aws --region us-east-1 --output my-infrastructure.json
```

### Discover Specific Resources

```bash
# Only VPCs
./bin/chimera discover --provider aws --region us-east-1 --resource-type vpc

# Only EC2 instances
./bin/chimera discover --provider aws --region us-east-1 --resource-type instance

# Multiple resource types
./bin/chimera discover --provider aws --region us-east-1 --resource-type vpc --resource-type subnet
```

### Multi-Region Discovery

```bash
# Scan multiple regions
./bin/chimera discover --provider aws --region us-east-1 --region us-west-2 --format table
```

## 🛠️ Development Commands

### Build and Test

```bash
# Clean rebuild
make clean && make build

# Run integration tests
make integration-test

# Verify Phase 1 completion
make phase1-test

# Format code
make fmt
```

### Configuration Management

```bash
# Create configuration file
./bin/chimera config init

# Validate configuration
./bin/chimera config validate

# Show current configuration
./bin/chimera config show
```

## 🐛 Troubleshooting

### Issue: "AWS credential validation failed"

**Solution 1: Check AWS CLI**
```bash
aws sts get-caller-identity
# If this fails, fix your AWS CLI setup first
```

**Solution 2: Set AWS Profile**
```bash
export AWS_PROFILE=your-profile-name
./bin/chimera discover --provider aws --region us-east-1
```

**Solution 3: Check Permissions**
Ensure your AWS credentials have these minimum permissions:
- `ec2:DescribeVpcs`
- `ec2:DescribeSubnets`
- `ec2:DescribeSecurityGroups`
- `ec2:DescribeInstances`
- `ec2:DescribeRegions`

### Issue: "No resources found"

This is normal if your AWS account/region has no resources. Try:
```bash
# Different region
./bin/chimera discover --provider aws --region us-west-2

# Dry run to see what would be scanned
./bin/chimera discover --provider aws --region us-east-1 --dry-run
```

### Issue: Build fails

```bash
# Clean dependencies and rebuild
go clean -modcache
make clean
make setup
make build
```

## 🔍 What Resources Are Discovered

Currently supported AWS resources:

| Resource Type | Description | Details Captured |
|---------------|-------------|------------------|
| **VPCs** | Virtual Private Clouds | CIDR blocks, state, default status |
| **Subnets** | VPC subnets | Availability zones, IP counts, public IP mapping |
| **Security Groups** | Network security rules | Ingress/egress rule counts, descriptions |
| **EC2 Instances** | Virtual machines | Instance types, states, IPs, launch times |

## 🚀 Next Steps

### 1. Explore Advanced Features

```bash
# Get help for any command
./bin/chimera discover --help
./bin/chimera config --help

# Run comprehensive demo
make demo

# Check project status
make status
```

### 2. Ready for Development

Your Phase 1 setup is complete! You now have:
- ✅ Working AWS discovery
- ✅ Professional CLI framework
- ✅ Configuration management
- ✅ Development environment

### 3. Contributing

Ready to add more cloud providers or features?
```bash
# Check what needs to be done
make help

# See the full roadmap in README.md
cat README.md
```

## 📊 Performance Expectations

- **Small AWS environments** (1-10 resources): < 2 seconds
- **Medium environments** (10-100 resources): 2-5 seconds  
- **Large environments** (100+ resources): 5-15 seconds

Discovery time scales with the number of resources and AWS API response times.

## 🎯 Architecture You've Built

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  AWS Account    │    │     Chimera      │    │  JSON/Table     │
│                 │───▶│   Discovery      │───▶│   Output        │
│ VPCs, Subnets   │    │   Engine         │    │                 │
│ SGs, Instances  │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

You've successfully built a production-ready infrastructure discovery tool!

## 📞 Getting Help

- **📚 Main Documentation**: [README.md](README.md)
- **🐛 Issues**: [GitHub Issues](https://github.com/BigChiefRick/chimera/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/BigChiefRick/chimera/discussions)

---

**🎉 Congratulations!** You've successfully set up Chimera Phase 1 with real AWS discovery.

**Ready for Phase 2?** Check out the roadmap in [README.md](README.md) to add Azure and GCP support!
