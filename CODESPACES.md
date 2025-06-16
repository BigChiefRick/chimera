# Chimera on GitHub Codespaces 🚀

The easiest way to develop Chimera is using GitHub Codespaces. Everything is pre-configured and ready to go!

## 🌟 Quick Start

### 1. Create Your Codespace

1. Go to [github.com/BigChiefRick/chimera](https://github.com/BigChiefRick/chimera)
2. Click the **"Code"** dropdown
3. Click **"Create codespace on main"**
4. Wait 2-3 minutes for setup to complete

### 2. Initialize Development Environment

Once your Codespace loads:

```bash
# Set up the development environment
make setup

# Start Chimera development tools
./scripts/codespaces.sh start
```

### 3. Configure Cloud Providers

#### AWS Setup
```bash
# Set up AWS credentials as Codespace secrets (recommended)
./scripts/codespaces.sh setup-aws

# OR use temporary credentials
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key" 
export AWS_DEFAULT_REGION="us-east-1"
```

#### Azure Setup
```bash
# Interactive device login (recommended for Codespaces)
az login --use-device-code
```

#### GCP Setup
```bash
# Interactive browser login
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### 4. Test Everything Works

```bash
# Run the comprehensive demo
./scripts/codespaces.sh demo

# Test individual cloud discoveries
make test-discovery-aws
make test-discovery-azure
make test-discovery-gcp
```

## 🛠️ Pre-installed Tools

Your Codespace comes with everything needed:

- **Go 1.21+** - Primary development language
- **Terraform** - Infrastructure as Code engine
- **Steampipe** - Multi-cloud SQL interface (auto-configured)
- **Terraformer** - Reverse engineering tool (latest version)
- **AWS CLI** - Amazon Web Services
- **Azure CLI** - Microsoft Azure
- **Google Cloud CLI** - Google Cloud Platform
- **Docker** - Containerization support
- **All Go dev tools** - gopls, golangci-lint, etc.

## 🔧 Development Workflow

### VS Code Integration

Your Codespace includes pre-configured VS Code tasks:

- **Ctrl+Shift+P** → **"Tasks: Run Task"**:
  - **Chimera: Build** - Build the project
  - **Chimera: Test** - Run tests
  - **Chimera: Integration Test** - Full integration test
  - **Steampipe: Start Service** - Start Steampipe
  - **Chimera: Quick Demo** - Test everything

### Command Line Development

```bash
# Development cycle
make fmt              # Format code
make vet              # Check for issues
make test             # Run unit tests
make build            # Build binary
make integration-test # Full integration test

# Discovery testing
make test-discovery-aws    # Test AWS discovery
make test-discovery-azure  # Test Azure discovery
make test-discovery-gcp    # Test GCP discovery
make test-discovery-all    # Test all clouds

# Debugging
./bin/chimera --help       # Test built binary
go run cmd/main.go --help  # Run from source
```

## 🌐 Port Forwarding

These ports are automatically forwarded to your local machine:

| Port | Service | Purpose |
|------|---------|---------|
| 9193 | Steampipe | PostgreSQL interface for cloud queries |
| 8080 | Chimera API | Future Chimera web interface |
| 3000 | Dev Server | Development server (when needed) |

Access them via: `https://CODESPACE-9193.app.github.dev` (replace with your Codespace URL)

## 💾 Persistent Storage

These directories persist across Codespace rebuilds:

- **Go modules** (`/go/pkg/mod`) - Dependencies cache
- **Steampipe** (`~/.steampipe`) - Configuration and plugin cache

Your code changes are always persistent in the Codespace.

## 🔐 Credential Management

### Recommended: Codespace Secrets

1. Go to your repository → **Settings** → **Secrets and variables** → **Codespaces**
2. Add these secrets:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_DEFAULT_REGION`

### Alternative: Environment Variables

```bash
# Set temporarily in your Codespace session
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_DEFAULT_REGION="us-east-1"

# Azure - use device login (easier in Codespaces)
az login --use-device-code

# GCP - use browser login
gcloud auth login
```

## 🚀 Example Discovery Session

```bash
# Start your development session
./scripts/codespaces.sh start

# Query AWS VPCs
steampipe query "
  select name, vpc_id, cidr_block, region 
  from aws_vpc 
  order by region
"

# Cross-cloud resource count
steampipe query "
  select 'AWS VPCs' as resource_type, count(*) as count from aws_vpc
  union all
  select 'Azure RGs' as resource_type, count(*) from azure_resource_group
  union all  
  select 'GCP Projects' as resource_type, count(*) from gcp_project
"

# Generate Terraform from existing AWS VPC
mkdir test-generation
cd test-generation
terraformer import aws --resources=vpc --regions=us-east-1
cat aws/vpc/vpc.tf
```

## 🎯 Machine Specifications

### Recommended Codespace Specs:

| Use Case | Machine Type | Cores | RAM | Cost/Hour* |
|----------|--------------|-------|-----|------------|
| **Development** | 2-core | 2 | 8GB | $0.18 |
| **Large Discovery** | 4-core | 4 | 16GB | $0.36 |
| **Heavy Testing** | 8-core | 8 | 32GB | $0.72 |

*Costs shown are GitHub's standard rates

### For Getting Started:
**2-core machine** is perfect for development and small-scale testing.

### For Production Testing:
**4-core machine** recommended when testing large cloud environments.

## 🆘 Troubleshooting

### Steampipe Issues
```bash
# Check Steampipe status
steampipe service status

# Restart Steampipe
steampipe service restart

# Test connection
steampipe query "select 'test' as message"

# Check plugins
steampipe plugin list
```

### Cloud Authentication Issues
```bash
# AWS
aws sts get-caller-identity

# Azure  
az account show

# GCP
gcloud auth list
gcloud config list
```

### Go/Build Issues
```bash
# Clean and rebuild
make clean
go mod tidy
make build

# Reset Go modules
go clean -modcache
go mod download
```

### Codespace Not Working?
```bash
# Re-run setup
.devcontainer/setup.sh

# Or rebuild the Codespace:
# Ctrl+Shift+P → "Codespaces: Rebuild Container"
```

## 🎉 Success Indicators

You'll know everything is working when:

✅ `make setup` completes successfully  
✅ `./scripts/codespaces.sh demo` passes all tests  
✅ `steampipe query "select 'test' as message"` returns result  
✅ `terraformer version` shows version info  
✅ `make test-discovery-aws` returns your VPCs (if AWS configured)  

## 💡 Pro Tips

1. **Use VS Code tasks** - Press `Ctrl+Shift+P` and type "Tasks" for quick actions

2. **Terminal shortcuts** - Use the integrated terminal (Ctrl+`) for all commands

3. **Port forwarding** - Access Steampipe externally via the forwarded port for SQL clients

4. **Codespace secrets** - Store credentials as secrets instead of environment variables

5. **Machine sizing** - Start with 2-core, upgrade to 4-core if working with large environments

6. **Prebuilds** - Enable prebuilds in your repo settings for faster Codespace startup

## 🔗 Useful Links

- **GitHub Codespaces Docs**: [docs.github.com/codespaces](https://docs.github.com/en/codespaces)
- **Steampipe Docs**: [steampipe.io/docs](https://steampipe.io/docs)
- **Terraformer Docs**: [github.com/GoogleCloudPlatform/terraformer](https://github.com/GoogleCloudPlatform/terraformer)
- **Chimera Repo**: [github.com/BigChiefRick/chimera](https://github.com/BigChiefRick/chimera)

---

**Ready to start? Create your Codespace and run `make setup`!** 🎯
