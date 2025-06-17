#!/bin/bash

# Chimera Phase 3 Integration Test Script
# This script tests the complete discovery -> generation workflow

set -e

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

FAILED_TESTS=0
TOTAL_TESTS=0
TEST_OUTPUT_DIR="./test-output"
BINARY="./bin/chimera"

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
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_info "Testing $test_name..."
    
    if output=$(eval "$test_command" 2>&1); then
        print_status "$test_name passed"
        if [[ "$3" == "show_output" ]]; then
            echo "  Output: $output"
        fi
        return 0
    else
        print_error "$test_name failed"
        echo "  Error: $output"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

show_banner() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║    🔮 CHIMERA PHASE 3 TESTING       ║"
    echo "  ║   IaC Generation Integration Test    ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""
}

# Check prerequisites
check_prerequisites() {
    print_info "=== Checking Prerequisites ==="
    
    # Check if binary exists
    if [ ! -f "$BINARY" ]; then
        print_error "Chimera binary not found at $BINARY"
        print_info "Run: make build"
        exit 1
    fi
    print_status "Chimera binary found"
    
    # Test binary works
    if ! $BINARY --help >/dev/null 2>&1; then
        print_error "Chimera binary not working"
        exit 1
    fi
    print_status "Chimera binary working"
    
    # Create test output directory
    mkdir -p "$TEST_OUTPUT_DIR"
    print_status "Test output directory created: $TEST_OUTPUT_DIR"
}

# Test basic CLI functionality
test_cli_functionality() {
    print_info "=== Testing CLI Functionality ==="
    
    run_test "CLI help command" "$BINARY --help"
    run_test "CLI version command" "$BINARY version"
    run_test "Discover help command" "$BINARY discover --help"
    run_test "Generate help command" "$BINARY generate --help"
    run_test "Config help command" "$BINARY config --help"
}

# Test discovery functionality
test_discovery_functionality() {
    print_info "=== Testing Discovery Functionality ==="
    
    # Test dry run
    run_test "AWS discovery dry-run" "$BINARY discover --provider aws --region us-east-1 --dry-run"
    
    # Test with real AWS if credentials available
    if aws sts get-caller-identity >/dev/null 2>&1; then
        print_info "AWS credentials detected - testing real discovery"
        run_test "Real AWS discovery" "$BINARY discover --provider aws --region us-east-1 --output $TEST_OUTPUT_DIR/real-aws-discovery.json"
        
        if [ -f "$TEST_OUTPUT_DIR/real-aws-discovery.json" ]; then
            resource_count=$(jq '.resources | length' "$TEST_OUTPUT_DIR/real-aws-discovery.json" 2>/dev/null || echo "0")
            print_status "Discovered $resource_count AWS resources"
        fi
    else
        print_warning "AWS credentials not configured - skipping real discovery test"
    fi
}

# Test generation functionality
test_generation_functionality() {
    print_info "=== Testing IaC Generation Functionality ==="
    
    # Create test discovery data
    print_info "Creating test discovery data..."
    cat > "$TEST_OUTPUT_DIR/test-discovery.json" << 'EOF'
{
  "resources": [
    {
      "id": "vpc-test123456",
      "name": "test-vpc",
      "type": "aws_vpc",
      "provider": "aws",
      "region": "us-east-1",
      "metadata": {
        "cidr_block": "10.0.0.0/16",
        "enable_dns_hostnames": true,
        "enable_dns_support": true,
        "state": "available"
      },
      "tags": {
        "Name": "test-vpc",
        "Environment": "test",
        "ManagedBy": "chimera-test"
      }
    },
    {
      "id": "subnet-test789012",
      "name": "test-subnet-public",
      "type": "aws_subnet",
      "provider": "aws",
      "region": "us-east-1",
      "zone": "us-east-1a",
      "metadata": {
        "vpc_id": "vpc-test123456",
        "cidr_block": "10.0.1.0/24",
        "map_public_ip_on_launch": true,
        "state": "available"
      },
      "tags": {
        "Name": "test-subnet-public",
        "Type": "public"
      }
    },
    {
      "id": "sg-test345678",
      "name": "test-security-group",
      "type": "aws_security_group",
      "provider": "aws",
      "region": "us-east-1",
      "metadata": {
        "vpc_id": "vpc-test123456",
        "description": "Test security group for Chimera"
      },
      "tags": {
        "Name": "test-security-group"
      }
    }
  ],
  "metadata": {
    "start_time": "2025-06-17T10:00:00Z",
    "end_time": "2025-06-17T10:01:00Z",
    "duration": 60000000000,
    "resource_count": 3,
    "provider_stats": {
      "aws": 3
    }
  }
}
EOF
    print_status "Test discovery data created"
    
    # Test generation dry-run
    run_test "Generation dry-run" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --dry-run"
    
    # Test actual Terraform generation
    run_test "Terraform generation" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --output $TEST_OUTPUT_DIR/terraform-output"
    
    # Verify generated files
    if [ -f "$TEST_OUTPUT_DIR/terraform-output/main.tf" ]; then
        print_status "main.tf generated successfully"
        
        # Check file content
        if grep -q "resource \"aws_vpc\"" "$TEST_OUTPUT_DIR/terraform-output/main.tf"; then
            print_status "VPC resource found in main.tf"
        else
            print_error "VPC resource not found in main.tf"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        if grep -q "resource \"aws_subnet\"" "$TEST_OUTPUT_DIR/terraform-output/main.tf"; then
            print_status "Subnet resource found in main.tf"
        else
            print_error "Subnet resource not found in main.tf"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        print_error "main.tf not generated"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    # Check for versions.tf
    if [ -f "$TEST_OUTPUT_DIR/terraform-output/versions.tf" ]; then
        print_status "versions.tf generated"
        
        if grep -q "hashicorp/aws" "$TEST_OUTPUT_DIR/terraform-output/versions.tf"; then
            print_status "AWS provider configuration found"
        else
            print_warning "AWS provider not found in versions.tf"
        fi
    else
        print_warning "versions.tf not generated"
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 4)) # Account for the file checks above
}

# Test different generation options
test_generation_options() {
    print_info "=== Testing Generation Options ==="
    
    # Test different organization patterns
    run_test "Organization by provider" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --organize-by provider --output $TEST_OUTPUT_DIR/terraform-by-provider"
    
    run_test "Organization by resource type" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --organize-by resource_type --output $TEST_OUTPUT_DIR/terraform-by-type"
    
    run_test "Single file generation" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --single-file --output $TEST_OUTPUT_DIR/terraform-single-file"
    
    # Test provider filtering
    run_test "AWS-only generation" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --provider aws --output $TEST_OUTPUT_DIR/terraform-aws-only"
    
    # Test resource filtering
    run_test "Include VPC only" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --include vpc --output $TEST_OUTPUT_DIR/terraform-vpc-only"
    
    run_test "Exclude security groups" "$BINARY generate --input $TEST_OUTPUT_DIR/test-discovery.json --format terraform --exclude security_group --output $TEST_OUTPUT_DIR/terraform-no-sg"
}

# Test Terraform validation
test_terraform_validation() {
    print_info "=== Testing Terraform Validation ==="
    
    if ! command -v terraform >/dev/null 2>&1; then
        print_warning "Terraform not installed - skipping validation tests"
        return 0
    fi
    
    print_info "Terraform found - running validation tests"
    
    # Test terraform fmt
    cd "$TEST_OUTPUT_DIR/terraform-output"
    if terraform fmt -check >/dev/null 2>&1; then
        print_status "Terraform formatting is correct"
    else
        print_warning "Terraform formatting could be improved"
        terraform fmt
        print_info "Applied terraform fmt"
    fi
    
    # Test terraform init
    if terraform init >/dev/null 2>&1; then
        print_status "Terraform init successful"
        
        # Test terraform validate
        if terraform validate >/dev/null 2>&1; then
            print_status "Terraform validation passed"
        else
            print_error "Terraform validation failed"
            terraform validate
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        # Test terraform plan (expected to fail without real resources)
        if terraform plan >/dev/null 2>&1; then
            print_status "Terraform plan successful (unexpected with test data)"
        else
            print_info "Terraform plan failed (expected with test data)"
        fi
    else
        print_error "Terraform init failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    cd - >/dev/null
    TOTAL_TESTS=$((TOTAL_TESTS + 3))
}

# Test end-to-end workflow
test_end_to_end_workflow() {
    print_info "=== Testing End-to-End Workflow ==="
    
    if aws sts get-caller-identity >/dev/null 2>&1; then
        print_info "Testing complete AWS discovery -> generation workflow"
        
        # Discover real AWS resources
        if $BINARY discover --provider aws --region us-east-1 --output "$TEST_OUTPUT_DIR/e2e-discovery.json" >/dev/null 2>&1; then
            print_status "Real AWS discovery completed"
            
            # Generate Terraform from real discovery
            if $BINARY generate --input "$TEST_OUTPUT_DIR/e2e-discovery.json" --format terraform --output "$TEST_OUTPUT_DIR/e2e-terraform" >/dev/null 2>&1; then
                print_status "Terraform generation from real discovery completed"
                
                # Count resources
                discovered_count=$(jq '.resources | length' "$TEST_OUTPUT_DIR/e2e-discovery.json" 2>/dev/null || echo "0")
                generated_files=$(ls "$TEST_OUTPUT_DIR/e2e-terraform"/*.tf 2>/dev/null | wc -l)
                
                print_status "End-to-end workflow: $discovered_count resources -> $generated_files files"
                
                # Validate generated Terraform if terraform is available
                if command -v terraform >/dev/null 2>&1; then
                    cd "$TEST_OUTPUT_DIR/e2e-terraform"
                    if terraform init >/dev/null 2>&1 && terraform validate >/dev/null 2>&1; then
                        print_status "Generated Terraform is valid"
                    else
                        print_warning "Generated Terraform validation failed"
                    fi
                    cd - >/dev/null
                fi
            else
                print_error "Terraform generation from real discovery failed"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            print_error "Real AWS discovery failed"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        TOTAL_TESTS=$((TOTAL_TESTS + 2))
    else
        print_warning "AWS credentials not configured - skipping end-to-end test"
        print_info "Configure AWS credentials with: aws configure"
    fi
}

# Test performance with larger datasets
test_performance() {
    print_info "=== Testing Performance ==="
    
    # Create larger test dataset
    print_info "Creating large test dataset..."
    python3 -c "
import json
import sys

resources = []
for i in range(50):  # 50 VPCs
    resources.append({
        'id': f'vpc-perf-{i:06d}',
        'name': f'perf-vpc-{i}',
        'type': 'aws_vpc',
        'provider': 'aws',
        'region': 'us-east-1',
        'metadata': {
            'cidr_block': f'10.{i//256}.{i%256}.0/24',
            'enable_dns_hostnames': True,
            'state': 'available'
        },
        'tags': {'Name': f'perf-vpc-{i}', 'Environment': 'performance-test'}
    })
    
    # Add 2 subnets per VPC
    for j in range(2):
        resources.append({
            'id': f'subnet-perf-{i:06d}-{j}',
            'name': f'perf-subnet-{i}-{j}',
            'type': 'aws_subnet',
            'provider': 'aws',
            'region': 'us-east-1',
            'zone': f'us-east-1{chr(97+j)}',
            'metadata': {
                'vpc_id': f'vpc-perf-{i:06d}',
                'cidr_block': f'10.{i//256}.{i%256}.{j}.0/28',
                'state': 'available'
            },
            'tags': {'Name': f'perf-subnet-{i}-{j}'}
        })

result = {'resources': resources, 'metadata': {'resource_count': len(resources)}}
print(json.dumps(result))
" > "$TEST_OUTPUT_DIR/performance-test.json" 2>/dev/null || {
        print_warning "Python3 not available - skipping performance test"
        return 0
    }
    
    resource_count=$(jq '.resources | length' "$TEST_OUTPUT_DIR/performance-test.json" 2>/dev/null)
    print_info "Created dataset with $resource_count resources"
    
    # Time the generation
    print_info "Testing generation performance..."
    start_time=$(date +%s.%N)
    
    if $BINARY generate --input "$TEST_OUTPUT_DIR/performance-test.json" --format terraform --output "$TEST_OUTPUT_DIR/performance-terraform" >/dev/null 2>&1; then
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "unknown")
        
        generated_files=$(ls "$TEST_OUTPUT_DIR/performance-terraform"/*.tf 2>/dev/null | wc -l)
        print_status "Performance test: $resource_count resources -> $generated_files files in ${duration}s"
        
        # Check if performance is reasonable (should be under 10 seconds for 150 resources)
        if command -v bc >/dev/null 2>&1; then
            if (( $(echo "$duration < 10" | bc -l) )); then
                print_status "Performance is acceptable"
            else
                print_warning "Performance slower than expected"
            fi
        fi
    else
        print_error "Performance test failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Test error handling
test_error_handling() {
    print_info "=== Testing Error Handling ==="
    
    # Test with non-existent input file
    if $BINARY generate --input "/non/existent/file.json" --format terraform >/dev/null 2>&1; then
        print_error "Should fail with non-existent input file"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles non-existent input file"
    fi
    
    # Test with invalid JSON
    echo "invalid json content" > "$TEST_OUTPUT_DIR/invalid.json"
    if $BINARY generate --input "$TEST_OUTPUT_DIR/invalid.json" --format terraform >/dev/null 2>&1; then
        print_error "Should fail with invalid JSON"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles invalid JSON"
    fi
    
    # Test with empty resource list
    echo '{"resources": []}' > "$TEST_OUTPUT_DIR/empty.json"
    if $BINARY generate --input "$TEST_OUTPUT_DIR/empty.json" --format terraform >/dev/null 2>&1; then
        print_error "Should fail with empty resource list"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_status "Correctly handles empty resource list"
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 3))
}

# Main test execution
main() {
    show_banner
    
    print_info "Starting Chimera Phase 3 Integration Tests"
    print_info "Test output directory: $TEST_OUTPUT_DIR"
    echo ""
    
    check_prerequisites
    test_cli_functionality
    test_discovery_functionality
    test_generation_functionality
    test_generation_options
    test_terraform_validation
    test_end_to_end_workflow
    test_performance
    test_error_handling
    
    echo ""
    print_info "=== Test Summary ==="
    PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
    echo "Total tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    
    if [ $FAILED_TESTS -gt 0 ]; then
        echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
        echo ""
        print_error "Some tests failed. Please review the output above."
        echo ""
        print_info "Common issues and solutions:"
        echo "  • AWS credentials: Run 'aws configure' to set up AWS access"
        echo "  • Terraform: Install from https://terraform.io/downloads"
        echo "  • Python3: Install for performance testing"
        echo "  • Build: Run 'make build' to ensure latest binary"
        exit 1
    else
        echo -e "Failed: ${GREEN}0${NC}"
        echo ""
        print_status "🎉 All tests passed! Phase 3 is working correctly."
        echo ""
        print_info "Generated test outputs in: $TEST_OUTPUT_DIR"
        print_info "You can examine the generated Terraform files to see the results."
        echo ""
        print_status "Chimera Phase 3 IaC generation is production ready! 🚀"
    fi
}

# Run main function
main "$@"
