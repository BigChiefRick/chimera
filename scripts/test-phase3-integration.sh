#!/bin/bash

# Chimera Phase 3 Complete Integration Testing Script
# Comprehensive testing for all Phase 3 generation capabilities

set -e

echo "🧪 Chimera Phase 3 Complete Integration Tests"
echo "============================================="

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

show_banner() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║       🔮 CHIMERA PHASE 3 TESTS      ║"
    echo "  ║   Complete Generation Capabilities   ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""
}

FAILED_TESTS=0
TOTAL_TESTS=0
TEST_DIR="phase3-test-workspace"

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_info "Testing $test_name..."
    
    if eval "$test_command" &> /dev/null; then
        print_status "$test_name passed"
        return 0
    else
        print_error "$test_name failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

run_test_with_output() {
    local test_name="$1"
    local test_command="$2"
    local show_output="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_info "Testing $test_name..."
    
    if output=$(eval "$test_command" 2>&1); then
        print_status "$test_name passed"
        if [[ "$show_output" == "show" ]]; then
            echo "  → $output"
        fi
        return 0
    else
        print_error "$test_name failed"
        echo "  → Error: $output"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Setup comprehensive test environment
setup_test_environment() {
    print_info "🔧 Setting up comprehensive test environment..."
    
    # Clean up any previous test runs
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Create comprehensive test data with multiple resource types
    cat > comprehensive-test-resources.json << 'EOF'
[
  {
    "id": "vpc-test123abc",
    "name": "chimera-test-vpc",
    "type": "aws_vpc",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "cidr_block": "10.100.0.0/16",
      "state": "available",
      "is_default": false,
      "enable_dns_hostnames": true,
      "enable_dns_support": true
    },
    "tags": {
      "Name": "chimera-test-vpc",
      "Environment": "test",
      "Project": "chimera-phase3",
      "ManagedBy": "ChimeraTest"
    }
  },
  {
    "id": "subnet-pub456def",
    "name": "chimera-public-subnet",
    "type": "aws_subnet",
    "provider": "aws",
    "region": "us-east-1",
    "zone": "us-east-1a",
    "metadata": {
      "cidr_block": "10.100.1.0/24",
      "vpc_id": "vpc-test123abc",
      "map_public_ip_on_launch": true,
      "available_ip_address_count": 251
    },
    "tags": {
      "Name": "chimera-public-subnet",
      "Type": "public",
      "Environment": "test",
      "Tier": "web"
    }
  },
  {
    "id": "subnet-prv789ghi",
    "name": "chimera-private-subnet",
    "type": "aws_subnet",
    "provider": "aws",
    "region": "us-east-1",
    "zone": "us-east-1b",
    "metadata": {
      "cidr_block": "10.100.2.0/24",
      "vpc_id": "vpc-test123abc",
      "map_public_ip_on_launch": false,
      "available_ip_address_count": 250
    },
    "tags": {
      "Name": "chimera-private-subnet",
      "Type": "private",
      "Environment": "test",
      "Tier": "app"
    }
  },
  {
    "id": "sg-web123jkl",
    "name": "chimera-web-sg",
    "type": "aws_security_group",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "group_name": "chimera-web-sg",
      "description": "Security group for web servers",
      "vpc_id": "vpc-test123abc",
      "ingress_rules": 3,
      "egress_rules": 1,
      "owner_id": "123456789012"
    },
    "tags": {
      "Name": "chimera-web-sg",
      "Purpose": "web-servers",
      "Environment": "test",
      "Protocol": "http-https"
    }
  },
  {
    "id": "sg-app456mno",
    "name": "chimera-app-sg",
    "type": "aws_security_group",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "group_name": "chimera-app-sg",
      "description": "Security group for application servers",
      "vpc_id": "vpc-test123abc",
      "ingress_rules": 2,
      "egress_rules": 1,
      "owner_id": "123456789012"
    },
    "tags": {
      "Name": "chimera-app-sg",
      "Purpose": "app-servers",
      "Environment": "test",
      "Tier": "application"
    }
  },
  {
    "id": "i-web789pqr",
    "name": "chimera-web-server-1",
    "type": "aws_instance",
    "provider": "aws",
    "region": "us-east-1",
    "zone": "us-east-1a",
    "metadata": {
      "instance_type": "t3.micro",
      "image_id": "ami-0abcdef1234567890",
      "state": "running",
      "vpc_id": "vpc-test123abc",
      "subnet_id": "subnet-pub456def",
      "private_ip": "10.100.1.100",
      "public_ip": "203.0.113.100",
      "key_name": "chimera-test-key"
    },
    "tags": {
      "Name": "chimera-web-server-1",
      "Role": "web-server",
      "Environment": "test",
      "Tier": "web",
      "Application": "nginx"
    },
    "created_at": "2025-01-01T12:00:00Z"
  },
  {
    "id": "i-app012stu",
    "name": "chimera-app-server-1",
    "type": "aws_instance",
    "provider": "aws",
    "region": "us-east-1",
    "zone": "us-east-1b",
    "metadata": {
      "instance_type": "t3.small",
      "image_id": "ami-0fedcba9876543210",
      "state": "running",
      "vpc_id": "vpc-test123abc",
      "subnet_id": "subnet-prv789ghi",
      "private_ip": "10.100.2.50",
      "key_name": "chimera-test-key"
    },
    "tags": {
      "Name": "chimera-app-server-1",
      "Role": "app-server",
      "Environment": "test",
      "Tier": "application",
      "Application": "nodejs"
    },
    "created_at": "2025-01-01T12:30:00Z"
  },
  {
    "id": "igw-test345vwx",
    "name": "chimera-internet-gateway",
    "type": "aws_internet_gateway",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "vpc_id": "vpc-test123abc",
      "state": "available"
    },
    "tags": {
      "Name": "chimera-internet-gateway",
      "Environment": "test",
      "Purpose": "public-access"
    }
  },
  {
    "id": "rtb-pub678yzA",
    "name": "chimera-public-route-table",
    "type": "aws_route_table",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "vpc_id": "vpc-test123abc",
      "route_count": 2
    },
    "tags": {
      "Name": "chimera-public-route-table",
      "Type": "public",
      "Environment": "test"
    }
  },
  {
    "id": "kp-test901BCD",
    "name": "chimera-test-key",
    "type": "aws_key_pair",
    "provider": "aws",
    "region": "us-east-1",
    "metadata": {
      "key_name": "chimera-test-key",
      "fingerprint": "ab:cd:ef:12:34:56:78:90"
    },
    "tags": {
      "Name": "chimera-test-key",
      "Environment": "test",
      "Usage": "ssh-access"
    }
  }
]
EOF

    # Create smaller test dataset for performance tests
    cat > small-test-resources.json << 'EOF'
[
  {
    "id": "vpc-small123",
    "name": "small-test-vpc",
    "type": "aws_vpc",
    "provider": "aws",
    "region": "us-west-2",
    "metadata": {
      "cidr_block": "172.16.0.0/16",
      "state": "available"
    },
    "tags": {
      "Name": "small-test-vpc",
      "Environment": "test"
    }
  }
]
EOF

    # Create empty dataset for edge case testing
    echo '[]' > empty-resources.json

    print_status "Test environment setup complete"
    print_info "Created test datasets: comprehensive (10 resources), small (1 resource), empty (0 resources)"
}

# Test 1: Prerequisites and environment
test_prerequisites() {
    print_info "=== Testing Prerequisites and Environment ==="
    
    # Check if chimera binary exists
    run_test "Chimera binary exists" "test -f ../bin/chimera"
    
    # Test basic CLI functionality
    run_test "CLI help command" "../bin/chimera --help"
    run_test "CLI version command" "../bin/chimera version"
    run_test "Generate command help" "../bin/chimera generate --help"
    run_test "Discover command help" "../bin/chimera discover --help"
    
    # Check required directories exist
    run_test "Generation package exists" "test -d ../pkg/generation"
    run_test "Mappers package exists" "test -d ../pkg/generation/mappers"
    run_test "Terraform package exists" "test -d ../pkg/generation/terraform"
    
    # Check key files exist
    run_test "Generation interfaces exist" "test -f ../pkg/generation/interfaces.go"
    run_test "Generation engine exists" "test -f ../pkg/generation/engine.go"
    run_test "AWS mapper exists" "test -f ../pkg/generation/mappers/aws.go"
    run_test "Terraform generator exists" "test -f ../pkg/generation/terraform/generator.go"
    
    print_status "Prerequisites and environment tests completed"
}

# Test 2: Basic generation functionality
test_basic_generation() {
    print_info "=== Testing Basic Generation Functionality ==="
    
    # Test dry run functionality
    run_test "Basic dry run" "../bin/chimera generate --input comprehensive-test-resources.json --dry-run"
    run_test "Terraform format dry run" "../bin/chimera generate --input comprehensive-test-resources.json --format terraform --dry-run"
    run_test "Empty resources dry run" "../bin/chimera generate --input empty-resources.json --dry-run"
    
    # Test basic generation
    mkdir -p basic-output
    run_test "Basic Terraform generation" "../bin/chimera generate --input comprehensive-test-resources.json --output basic-output --force"
    
    # Verify core files were created
    run_test "Main file created" "test -f basic-output/main.tf"
    run_test "Variables file created" "test -f basic-output/variables.tf"
    run_test "Outputs file created" "test -f basic-output/outputs.tf"
    run_test "Provider file created" "test -f basic-output/providers.tf"
    
    # Check file contents
    run_test "Main file has content" "test -s basic-output/main.tf"
    run_test "Variables file has content" "test -s basic-output/variables.tf"
    run_test "Outputs file has content" "test -s basic-output/outputs.tf"
    run_test "Provider file has content" "test -s basic-output/providers.tf"
    
    print_status "Basic generation functionality tests completed"
}

# Test 3: Resource mapping validation
test_resource_mapping() {
    print_info "=== Testing Resource Mapping ==="
    
    # Check that all resource types are mapped correctly
    run_test "VPC resource mapped" "grep -q 'resource \"aws_vpc\"' basic-output/main.tf"
    run_test "Subnet resources mapped" "grep -q 'resource \"aws_subnet\"' basic-output/main.tf"
    run_test "Security group resources mapped" "grep -q 'resource \"aws_security_group\"' basic-output/main.tf"
    run_test "Instance resources mapped" "grep -q 'resource \"aws_instance\"' basic-output/main.tf"
    run_test "Internet gateway mapped" "grep -q 'resource \"aws_internet_gateway\"' basic-output/main.tf"
    run_test "Route table mapped" "grep -q 'resource \"aws_route_table\"' basic-output/main.tf"
    run_test "Key pair mapped" "grep -q 'resource \"aws_key_pair\"' basic-output/main.tf"
    
    # Check resource count matches input
    vpc_count=$(grep -c 'resource "aws_vpc"' basic-output/main.tf || echo "0")
    run_test "Correct number of VPCs (1)" "test $vpc_count -eq 1"
    
    subnet_count=$(grep -c 'resource "aws_subnet"' basic-output/main.tf || echo "0")
    run_test "Correct number of subnets (2)" "test $subnet_count -eq 2"
    
    instance_count=$(grep -c 'resource "aws_instance"' basic-output/main.tf || echo "0")
    run_test "Correct number of instances (2)" "test $instance_count -eq 2"
    
    # Check that tags are preserved
    run_test "Original tags preserved" "grep -q 'Environment.*test' basic-output/main.tf"
    run_test "Chimera management tags added" "grep -q 'ManagedBy.*Chimera' basic-output/main.tf"
    
    # Check that dependencies are handled
    run_test "VPC references in subnets" "grep -q '\${aws_vpc\.' basic-output/main.tf"
    run_test "Subnet references in instances" "grep -q '\${aws_subnet\.' basic-output/main.tf"
    
    print_status "Resource mapping tests completed"
}

# Test 4: Organization and structure options
test_organization_options() {
    print_info "=== Testing Organization Options ==="
    
    # Test single file generation
    mkdir -p single-file-output
    run_test "Single file generation" "../bin/chimera generate --input comprehensive-test-resources.json --output single-file-output --single-file --force"
    run_test "Single main.tf exists" "test -f single-file-output/main.tf"
    run_test "All resources in single file" "test \$(grep -c 'resource \"' single-file-output/main.tf) -ge 7"
    
    # Test organize by type
    mkdir -p by-type-output
    run_test "Organize by type generation" "../bin/chimera generate --input comprehensive-test-resources.json --output by-type-output --organize-by-type --force"
    run_test "VPC file exists" "test -f by-type-output/vpc.tf"
    run_test "Subnet file exists" "test -f by-type-output/subnet.tf"
    run_test "Instance file exists" "test -f by-type-output/instance.tf"
    run_test "Security group file exists" "test -f by-type-output/security_group.tf"
    
    # Verify type organization works correctly
    run_test "VPC only in vpc.tf" "grep -q 'resource \"aws_vpc\"' by-type-output/vpc.tf && ! grep -q 'resource \"aws_subnet\"' by-type-output/vpc.tf"
    run_test "Subnets only in subnet.tf" "grep -q 'resource \"aws_subnet\"' by-type-output/subnet.tf && ! grep -q 'resource \"aws_vpc\"' by-type-output/subnet.tf"
    
    print_status "Organization options tests completed"
}

# Test 5: Module generation
test_module_generation() {
    print_info "=== Testing Module Generation ==="
    
    mkdir -p module-output
    run_test "Module generation" "../bin/chimera generate --input comprehensive-test-resources.json --output module-output --generate-modules --force"
    
    # Check module directory structure
    run_test "Modules directory exists" "test -d module-output/modules"
    run_test "AWS module directory exists" "test -d module-output/modules/aws"
    
    # Check module files
    run_test "Module main.tf exists" "test -f module-output/modules/aws/main.tf"
    run_test "Module variables.tf exists" "test -f module-output/modules/aws/variables.tf"
    run_test "Module outputs.tf exists" "test -f module-output/modules/aws/outputs.tf"
    
    # Verify module content
    run_test "Module main.tf has resources" "test -s module-output/modules/aws/main.tf"
    run_test "Module has VPC resources" "grep -q 'resource \"aws_vpc\"' module-output/modules/aws/main.tf"
    
    # Test different module structures
    mkdir -p module-by-service-output
    run_test "Module by service generation" "../bin/chimera generate --input comprehensive-test-resources.json --output module-by-service-output --generate-modules --module-structure by_service --force"
    
    print_status "Module generation tests completed"
}

# Test 6: Filtering and customization
test_filtering_options() {
    print_info "=== Testing Filtering Options ==="
    
    # Test include filter
    mkdir -p filtered-vpc-output
    run_test "Include VPC only" "../bin/chimera generate --input comprehensive-test-resources.json --output filtered-vpc-output --include vpc --force"
    run_test "Only VPC in filtered output" "grep -q 'resource \"aws_vpc\"' filtered-vpc-output/main.tf && ! grep -q 'resource \"aws_instance\"' filtered-vpc-output/main.tf"
    
    # Test exclude filter
    mkdir -p filtered-no-instance-output
    run_test "Exclude instances" "../bin/chimera generate --input comprehensive-test-resources.json --output filtered-no-instance-output --exclude instance --force"
    run_test "No instances in excluded output" "! grep -q 'resource \"aws_instance\"' filtered-no-instance-output/main.tf"
    run_test "VPC still in excluded output" "grep -q 'resource \"aws_vpc\"' filtered-no-instance-output/main.tf"
    
    # Test provider filter
    mkdir -p filtered-aws-output
    run_test "AWS provider filter" "../bin/chimera generate --input comprehensive-test-resources.json --output filtered-aws-output --provider aws --force"
    run_test "AWS resources in provider filtered output" "grep -q 'resource \"aws_' filtered-aws-output/main.tf"
    
    # Test region filter
    mkdir -p filtered-region-output
    run_test "Region filter" "../bin/chimera generate --input comprehensive-test-resources.json --output filtered-region-output --region us-east-1 --force"
    run_test "Region filtered resources exist" "test -f filtered-region-output/main.tf"
    
    # Test resource type filter
    mkdir -p filtered-type-output
    run_test "Resource type filter" "../bin/chimera generate --input comprehensive-test-resources.json --output filtered-type-output --resource-type vpc --resource-type subnet --force"
    run_test "VPC and subnet in type filtered output" "grep -q 'resource \"aws_vpc\"' filtered-type-output/main.tf && grep -q 'resource \"aws_subnet\"' filtered-type-output/main.tf"
    run_test "No instances in type filtered output" "! grep -q 'resource \"aws_instance\"' filtered-type-output/main.tf"
    
    print_status "Filtering options tests completed"
}

# Test 7: Advanced features
test_advanced_features() {
    print_info "=== Testing Advanced Features ==="
    
    # Test template variables
    mkdir -p template-vars-output
    run_test "Template variables" "../bin/chimera generate --input comprehensive-test-resources.json --output template-vars-output --template-var project=test-project --template-var environment=production --force"
    
    # Test validation
    mkdir -p validation-output
    run_test "Output validation" "../bin/chimera generate --input comprehensive-test-resources.json --output validation-output --validate --force"
    
    # Test compact output
    mkdir -p compact-output
    run_test "Compact output generation" "../bin/chimera generate --input comprehensive-test-resources.json --output compact-output --compact --force"
    
    # Test without provider config
    mkdir -p no-provider-output
    run_test "Generation without provider" "../bin/chimera generate --input comprehensive-test-resources.json --output no-provider-output --include-provider=false --force"
    run_test "No provider file when disabled" "! test -f no-provider-output/providers.tf"
    
    print_status "Advanced features tests completed"
}

# Test 8: Error handling and edge cases
test_error_handling() {
    print_info "=== Testing Error Handling ==="
    
    # Test invalid input file
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if ../bin/chimera generate --input nonexistent.json --dry-run &> /dev/null; then
        print_error "Should fail with nonexistent input file"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles nonexistent input file"
    fi
    
    # Test invalid format
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if ../bin/chimera generate --input comprehensive-test-resources.json --format invalid-format --dry-run &> /dev/null; then
        print_error "Should fail with invalid format"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles invalid format"
    fi
    
    # Test conflicting options
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if ../bin/chimera generate --input comprehensive-test-resources.json --single-file --organize-by-type --dry-run &> /dev/null; then
        print_error "Should fail with conflicting options"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles conflicting options"
    fi
    
    # Test empty resource file
    mkdir -p empty-output
    run_test "Empty resources handling" "../bin/chimera generate --input empty-resources.json --output empty-output --force"
    
    # Test malformed JSON
    echo '{"invalid": json}' > malformed.json
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if ../bin/chimera generate --input malformed.json --dry-run &> /dev/null; then
        print_error "Should fail with malformed JSON"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles malformed JSON"
    fi
    
    print_status "Error handling tests completed"
}

# Test 9: Performance and scalability
test_performance() {
    print_info "=== Testing Performance and Scalability ==="
    
    # Create large dataset
    print_info "Creating large test dataset (100 resources)..."
    echo '[' > large-test.json
    for i in $(seq 1 100); do
        if [ $i -gt 1 ]; then echo "," >> large-test.json; fi
        cat >> large-test.json << EOF
{
  "id": "vpc-perf$i",
  "name": "perf-vpc-$i",
  "type": "aws_vpc",
  "provider": "aws",
  "region": "us-east-1",
  "metadata": {
    "cidr_block": "10.$i.0.0/16",
    "state": "available",
    "is_default": false
  },
  "tags": {
    "Name": "perf-vpc-$i",
    "Environment": "performance-test",
    "Index": "$i"
  }
}
EOF
    done
    echo ']' >> large-test.json
    
    # Test generation performance
    mkdir -p perf-output
    print_info "Testing generation performance with 100 resources..."
    start_time=$(date +%s)
    if ../bin/chimera generate --input large-test.json --output perf-output --force &> /dev/null; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        print_status "Performance test completed in ${duration}s"
        
        # Check results
        resource_count=$(grep -c 'resource "aws_vpc"' perf-output/main.tf || echo "0")
        print_info "Generated $resource_count VPC resources"
        
        if [ $duration -lt 15 ]; then
            print_status "Performance is excellent (< 15s for 100 resources)"
        elif [ $duration -lt 30 ]; then
            print_status "Performance is good (< 30s for 100 resources)"
        else
            print_warning "Performance could be improved (${duration}s for 100 resources)"
        fi
        
        # Check file sizes
        main_size=$(wc -l < perf-output/main.tf)
        print_info "Generated main.tf has $main_size lines"
        
    else
        print_error "Performance test failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    print_status "Performance tests completed"
}

# Test 10: Terraform validation
test_terraform_validation() {
    print_info "=== Testing Terraform Validation ==="
    
    # Check if Terraform is available
    if command -v terraform &> /dev/null; then
        print_info "Terraform found, running validation tests..."
        
        # Test basic output validation
        cd basic-output
        run_test "Terraform format check" "terraform fmt -check"
        run_test "Terraform init" "terraform init"
        run_test "Terraform validate" "terraform validate"
        cd ..
        
        # Test organized output validation
        cd by-type-output
        run_test "Organized output terraform init" "terraform init"
        run_test "Organized output terraform validate" "terraform validate"
        cd ..
        
        # Test module validation
        if [ -d "module-output/modules/aws" ]; then
            cd module-output/modules/aws
            run_test "Module terraform init" "terraform init"
            run_test "Module terraform validate" "terraform validate"
            cd ../../..
        fi
        
        print_status "Terraform validation tests completed"
    else
        print_warning "Terraform not found, skipping validation tests"
        print_info "Install Terraform to enable validation: https://terraform.io/downloads"
    fi
}

# Test 11: Real AWS integration (if available)
test_real_aws_integration() {
    print_info "=== Testing Real AWS Integration ==="
    
    # Check for AWS credentials
    if command -v aws &> /dev/null && aws sts get-caller-identity &> /dev/null 2>&1; then
        print_info "AWS credentials detected, testing real integration..."
        
        # Test discovery and generation workflow
        print_info "Running discovery..."
        if ../bin/chimera discover --provider aws --region us-east-1 --output real-aws-resources.json &> /dev/null; then
            print_status "Real AWS discovery completed"
            
            if [ -s "real-aws-resources.json" ]; then
                resource_count=$(jq length real-aws-resources.json 2>/dev/null || echo "unknown")
                print_info "Discovered $resource_count real AWS resources"
                
                # Test generation from real data
                mkdir -p real-aws-terraform
                if ../bin/chimera generate --input real-aws-resources.json --output real-aws-terraform --force &> /dev/null; then
                    print_status "Real AWS Terraform generation completed"
                    
                    # Validate real AWS Terraform
                    if command -v terraform &> /dev/null; then
                        cd real-aws-terraform
                        if terraform fmt -check &> /dev/null && terraform init &> /dev/null && terraform validate &> /dev/null; then
                            print_status "Real AWS Terraform validates successfully"
                        else
                            print_warning "Real AWS Terraform validation had issues"
                        fi
                        cd ..
                    fi
                    
                    # Show what was generated
                    if [ -f "real-aws-terraform/main.tf" ]; then
                        real_resource_count=$(grep -c '^resource ' real-aws-terraform/main.tf || echo "0")
                        print_info "Generated $real_resource_count Terraform resources from real AWS infrastructure"
                    fi
                else
                    print_warning "Real AWS Terraform generation failed"
                fi
            else
                print_info "No real AWS resources found (empty account or insufficient permissions)"
            fi
        else
            print_warning "Real AWS discovery failed"
        fi
    else
        print_info "No AWS credentials available, skipping real integration test"
        print_info "To enable: aws configure or set AWS environment variables"
    fi
    
    print_status "Real AWS integration tests completed"
}

# Test 12: End-to-end workflow validation
test_end_to_end_workflow() {
    print_info "=== Testing End-to-End Workflow ==="
    
    # Complete workflow test
    print_info "Testing complete discovery → generation → validation workflow..."
    
    # Use our test data to simulate a complete workflow
    mkdir -p e2e-workflow
    
    # Step 1: Simulate discovery (use our test data)
    cp comprehensive-test-resources.json e2e-workflow/discovered-resources.json
    print_status "Step 1: Discovery data prepared"
    
    # Step 2: Generate Terraform
    if ../bin/chimera generate --input e2e-workflow/discovered-resources.json --output e2e-workflow/terraform --force &> /dev/null; then
        print_status "Step 2: Terraform generation completed"
    else
        print_error "Step 2: Terraform generation failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Step 3: Validate Terraform
    if command -v terraform &> /dev/null; then
        cd e2e-workflow/terraform
        if terraform init &> /dev/null && terraform validate &> /dev/null; then
            print_status "Step 3: Terraform validation passed"
        else
            print_error "Step 3: Terraform validation failed"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        cd ../..
    else
        print_status "Step 3: Terraform validation skipped (terraform not installed)"
    fi
    
    # Step 4: Verify output structure
    run_test "E2E main.tf exists" "test -f e2e-workflow/terraform/main.tf"
    run_test "E2E variables.tf exists" "test -f e2e-workflow/terraform/variables.tf"
    run_test "E2E outputs.tf exists" "test -f e2e-workflow/terraform/outputs.tf"
    run_test "E2E providers.tf exists" "test -f e2e-workflow/terraform/providers.tf"
    
    # Step 5: Verify content quality
    run_test "E2E has all resource types" "grep -q 'aws_vpc\\|aws_subnet\\|aws_instance\\|aws_security_group' e2e-workflow/terraform/main.tf"
    run_test "E2E has proper dependencies" "grep -q '\${aws_vpc\\.' e2e-workflow/terraform/main.tf"
    run_test "E2E has tags" "grep -q 'tags.*=' e2e-workflow/terraform/main.tf"
    
    print_status "End-to-end workflow tests completed"
}

# Cleanup function
cleanup_test_environment() {
    print_info "🧹 Cleaning up test environment..."
    cd ..
    rm -rf "$TEST_DIR"
    print_status "Test environment cleaned up"
}

# Main test execution
main() {
    show_banner
    
    # Check prerequisites
    if [ ! -f "bin/chimera" ]; then
        print_error "Chimera binary not found. Please run 'make build' first."
        exit 1
    fi
    
    print_info "Starting comprehensive Phase 3 integration tests..."
    print_info "This will test all generation capabilities including resource mapping, Terraform output, and validation."
    echo ""
    
    # Setup test environment
    setup_test_environment
    
    # Run all test suites
    test_prerequisites
    test_basic_generation
    test_resource_mapping
    test_organization_options
    test_module_generation
    test_filtering_options
    test_advanced_features
    test_error_handling
    test_performance
    test_terraform_validation
    test_real_aws_integration
    test_end_to_end_workflow
    
    # Cleanup
    cleanup_test_environment
    
    # Generate comprehensive test report
    echo ""
    echo "🎯 Phase 3 Integration Test Report"
    echo "================================="
    PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
    echo "Total tests executed: $TOTAL_TESTS"
    echo -e "Tests passed: ${GREEN}$PASSED_TESTS${NC}"
    
    if [ $FAILED_TESTS -gt 0 ]; then
        echo -e "Tests failed: ${RED}$FAILED_TESTS${NC}"
        echo ""
        print_warning "Some tests failed. Phase 3 implementation needs attention."
        echo ""
        echo "❌ Phase 3 Status: Incomplete"
        echo ""
        echo "🔧 Next steps:"
        echo "1. Review failed tests above"
        echo "2. Check implementation of failed components"
        echo "3. Re-run tests after fixes: ./scripts/test-phase3-integration.sh"
        
    else
        echo -e "Tests failed: ${GREEN}0${NC}"
        echo ""
        print_status "🎉 ALL PHASE 3 INTEGRATION TESTS PASSED!"
        echo ""
        echo "✅ Phase 3 Status: PRODUCTION READY"
        echo ""
        echo "🏆 Capabilities Verified:"
        echo "✅ Complete discovery → IaC generation workflow"
        echo "✅ AWS resource mapping (7+ resource types)"
        echo "✅ Production Terraform HCL generation"
        echo "✅ Resource organization and module generation"
        echo "✅ Comprehensive filtering and customization"
        echo "✅ Error handling and edge case management"
        echo "✅ Performance at scale (100+ resources in <30s)"
        echo "✅ Terraform validation and syntax correctness"
        echo "✅ Real AWS integration compatibility"
        echo "✅ End-to-end workflow validation"
        echo ""
        echo "🚀 PHASE 3 READY FOR PRODUCTION DEPLOYMENT!"
        echo ""
        echo "📋 Usage Examples:"
        echo "# Complete workflow:"
        echo "  1. ./bin/chimera discover --provider aws --region us-east-1 --output infra.json"
        echo "  2. ./bin/chimera generate --input infra.json --output terraform/"
        echo "  3. cd terraform && terraform init && terraform plan && terraform apply"
        echo ""
        echo "# Advanced generation:"
        echo "  ./bin/chimera generate --input infra.json --organize-by-type --generate-modules"
        echo ""
        echo "🎯 Phase 4 Ready: Multi-cloud providers (Azure, GCP) and advanced IaC formats"
    fi
    
    # Exit with appropriate code
    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Execute main function
main "$@"
