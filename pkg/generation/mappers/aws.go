package mappers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/BigChiefRick/chimera/pkg/discovery"
	"github.com/BigChiefRick/chimera/pkg/generation"
)

// AWSMapper implements ResourceMapper for AWS resources
type AWSMapper struct {
	resourceMappings map[string]string
	supportedTypes   []string
}

// NewAWSMapper creates a new AWS resource mapper
func NewAWSMapper() *AWSMapper {
	mapper := &AWSMapper{
		resourceMappings: map[string]string{
			"aws_vpc":            "aws_vpc",
			"aws_subnet":         "aws_subnet",
			"aws_security_group": "aws_security_group",
			"aws_instance":       "aws_instance",
			"aws_internet_gateway": "aws_internet_gateway",
			"aws_route_table":    "aws_route_table",
			"aws_nat_gateway":    "aws_nat_gateway",
			"aws_elastic_ip":     "aws_eip",
			"aws_network_acl":    "aws_network_acl",
		},
		supportedTypes: []string{
			"aws_vpc",
			"aws_subnet", 
			"aws_security_group",
			"aws_instance",
			"aws_internet_gateway",
			"aws_route_table",
			"aws_nat_gateway",
			"aws_elastic_ip",
			"aws_network_acl",
		},
	}
	return mapper
}

// Provider returns the cloud provider this mapper supports
func (m *AWSMapper) Provider() discovery.CloudProvider {
	return discovery.AWS
}

// GetSupportedTypes returns the resource types this mapper supports
func (m *AWSMapper) GetSupportedTypes() []string {
	return m.supportedTypes
}

// MapResource maps a discovered AWS resource to a Terraform resource
func (m *AWSMapper) MapResource(resource discovery.Resource) (*generation.TerraformResource, error) {
	if resource.Provider != discovery.AWS {
		return nil, fmt.Errorf("resource is not an AWS resource: %s", resource.Provider)
	}

	terraformType, exists := m.resourceMappings[resource.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported AWS resource type: %s", resource.Type)
	}

	// Generate clean Terraform resource name
	terraformName := m.generateResourceName(resource)

	// Map resource configuration based on type
	config, variables, outputs, dependencies, err := m.mapResourceConfig(resource, terraformType)
	if err != nil {
		return nil, fmt.Errorf("failed to map resource configuration: %w", err)
	}

	terraformResource := &generation.TerraformResource{
		Type:         terraformType,
		Name:         terraformName,
		Provider:     resource.Provider,
		Config:       config,
		Dependencies: dependencies,
		Outputs:      outputs,
		Variables:    variables,
		SourceInfo: generation.SourceInfo{
			OriginalID:       resource.ID,
			OriginalType:     resource.Type,
			OriginalProvider: resource.Provider,
			OriginalRegion:   resource.Region,
			DiscoveredAt:     time.Now(),
			Metadata:         resource.Metadata,
			Tags:             resource.Tags,
		},
	}

	return terraformResource, nil
}

// generateResourceName generates a clean Terraform resource name
func (m *AWSMapper) generateResourceName(resource discovery.Resource) string {
	// Start with resource name if available
	name := resource.Name
	if name == "" {
		// Use ID as fallback
		name = resource.ID
	}

	// Clean the name for Terraform
	name = m.cleanTerraformName(name)

	// If still empty, generate from resource type and ID
	if name == "" {
		resourceType := strings.TrimPrefix(resource.Type, "aws_")
		idSuffix := resource.ID
		if len(idSuffix) > 8 {
			idSuffix = idSuffix[len(idSuffix)-8:]
		}
		name = fmt.Sprintf("%s_%s", resourceType, idSuffix)
	}

	return name
}

// cleanTerraformName cleans a name for use in Terraform
func (m *AWSMapper) cleanTerraformName(name string) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	cleaned := reg.ReplaceAllString(name, "_")
	
	// Remove leading/trailing underscores
	cleaned = strings.Trim(cleaned, "_")
	
	// Ensure it starts with a letter or underscore
	if len(cleaned) > 0 && !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(cleaned) {
		cleaned = "resource_" + cleaned
	}
	
	// Convert to lowercase
	cleaned = strings.ToLower(cleaned)
	
	return cleaned
}

// mapResourceConfig maps the resource configuration based on resource type
func (m *AWSMapper) mapResourceConfig(resource discovery.Resource, terraformType string) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := make(map[string]interface{})
	variables := make(map[string]generation.Variable)
	outputs := make(map[string]string)
	var dependencies []string

	switch terraformType {
	case "aws_vpc":
		return m.mapVPC(resource)
	case "aws_subnet":
		return m.mapSubnet(resource)
	case "aws_security_group":
		return m.mapSecurityGroup(resource)
	case "aws_instance":
		return m.mapInstance(resource)
	case "aws_internet_gateway":
		return m.mapInternetGateway(resource)
	case "aws_route_table":
		return m.mapRouteTable(resource)
	case "aws_nat_gateway":
		return m.mapNATGateway(resource)
	case "aws_eip":
		return m.mapElasticIP(resource)
	case "aws_network_acl":
		return m.mapNetworkACL(resource)
	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported terraform resource type: %s", terraformType)
	}
}

// mapVPC maps AWS VPC resource
func (m *AWSMapper) mapVPC(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract CIDR block from metadata
	if cidrBlock, exists := resource.Metadata["cidr_block"]; exists {
		config["cidr_block"] = cidrBlock
	} else {
		return nil, nil, nil, nil, fmt.Errorf("VPC missing required cidr_block")
	}

	// Optional VPC settings
	if enableDnsHostnames, exists := resource.Metadata["enable_dns_hostnames"]; exists {
		config["enable_dns_hostnames"] = enableDnsHostnames
	}

	if enableDnsSupport, exists := resource.Metadata["enable_dns_support"]; exists {
		config["enable_dns_support"] = enableDnsSupport
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)):         "${aws_vpc." + m.generateResourceName(resource) + ".id}",
		fmt.Sprintf("%s_cidr_block", m.cleanTerraformName(resource.Name)): "${aws_vpc." + m.generateResourceName(resource) + ".cidr_block}",
	}

	return config, map[string]generation.Variable{}, outputs, []string{}, nil
}

// mapSubnet maps AWS Subnet resource
func (m *AWSMapper) mapSubnet(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract required fields
	vpcId, exists := resource.Metadata["vpc_id"]
	if !exists {
		return nil, nil, nil, nil, fmt.Errorf("subnet missing required vpc_id")
	}
	
	cidrBlock, exists := resource.Metadata["cidr_block"]
	if !exists {
		return nil, nil, nil, nil, fmt.Errorf("subnet missing required cidr_block")
	}

	config["vpc_id"] = fmt.Sprintf("${aws_vpc.%s.id}", m.cleanVPCReference(vpcId))
	config["cidr_block"] = cidrBlock
	config["availability_zone"] = resource.Zone

	// Optional subnet settings
	if mapPublicIp, exists := resource.Metadata["map_public_ip_on_launch"]; exists {
		config["map_public_ip_on_launch"] = mapPublicIp
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_subnet." + m.generateResourceName(resource) + ".id}",
	}

	// Add VPC dependency
	dependencies := []string{fmt.Sprintf("aws_vpc.%s", m.cleanVPCReference(vpcId))}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapSecurityGroup maps AWS Security Group resource
func (m *AWSMapper) mapSecurityGroup(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"name": resource.Name,
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract VPC ID
	if vpcId, exists := resource.Metadata["vpc_id"]; exists && vpcId != "" {
		config["vpc_id"] = fmt.Sprintf("${aws_vpc.%s.id}", m.cleanVPCReference(vpcId))
	}

	// Extract description
	if description, exists := resource.Metadata["description"]; exists {
		config["description"] = description
	}

	// Note: Actual ingress/egress rules would need to be discovered separately
	// For now, we'll add a comment about manual rule configuration
	
	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_security_group." + m.generateResourceName(resource) + ".id}",
	}

	var dependencies []string
	if vpcId, exists := resource.Metadata["vpc_id"]; exists && vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.cleanVPCReference(vpcId)))
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapInstance maps AWS EC2 Instance resource
func (m *AWSMapper) mapInstance(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract required fields
	instanceType, exists := resource.Metadata["instance_type"]
	if !exists {
		return nil, nil, nil, nil, fmt.Errorf("instance missing required instance_type")
	}
	config["instance_type"] = instanceType

	// AMI ID - Note: This may need to be parameterized
	if imageId, exists := resource.Metadata["image_id"]; exists {
		config["ami"] = imageId
	} else {
		// Create a variable for AMI since it might be region-specific
		variables := map[string]generation.Variable{
			"ami_id": {
				Name:        "ami_id",
				Type:        "string",
				Description: fmt.Sprintf("AMI ID for instance %s", resource.Name),
				Required:    true,
			},
		}
		config["ami"] = "${var.ami_id}"
		return config, variables, map[string]string{}, []string{}, nil
	}

	// Network configuration
	if subnetId, exists := resource.Metadata["subnet_id"]; exists && subnetId != "" {
		config["subnet_id"] = fmt.Sprintf("${aws_subnet.%s.id}", m.cleanSubnetReference(subnetId))
	}

	if vpcSecurityGroupIds, exists := resource.Metadata["vpc_security_group_ids"]; exists {
		// This would be an array of security group IDs
		config["vpc_security_group_ids"] = vpcSecurityGroupIds
	}

	// Optional instance settings
	if keyName, exists := resource.Metadata["key_name"]; exists && keyName != "" {
		config["key_name"] = keyName
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)):         "${aws_instance." + m.generateResourceName(resource) + ".id}",
		fmt.Sprintf("%s_private_ip", m.cleanTerraformName(resource.Name)): "${aws_instance." + m.generateResourceName(resource) + ".private_ip}",
	}

	if publicIp, exists := resource.Metadata["public_ip"]; exists && publicIp != "" {
		outputs[fmt.Sprintf("%s_public_ip", m.cleanTerraformName(resource.Name))] = "${aws_instance." + m.generateResourceName(resource) + ".public_ip}"
	}

	var dependencies []string
	if subnetId, exists := resource.Metadata["subnet_id"]; exists && subnetId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_subnet.%s", m.cleanSubnetReference(subnetId)))
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapInternetGateway maps AWS Internet Gateway resource
func (m *AWSMapper) mapInternetGateway(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract VPC ID if available
	var dependencies []string
	if vpcId, exists := resource.Metadata["vpc_id"]; exists && vpcId != "" {
		config["vpc_id"] = fmt.Sprintf("${aws_vpc.%s.id}", m.cleanVPCReference(vpcId))
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.cleanVPCReference(vpcId)))
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_internet_gateway." + m.generateResourceName(resource) + ".id}",
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapRouteTable maps AWS Route Table resource
func (m *AWSMapper) mapRouteTable(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract VPC ID
	var dependencies []string
	if vpcId, exists := resource.Metadata["vpc_id"]; exists && vpcId != "" {
		config["vpc_id"] = fmt.Sprintf("${aws_vpc.%s.id}", m.cleanVPCReference(vpcId))
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.cleanVPCReference(vpcId)))
	}

	// Note: Routes would need to be discovered and configured separately

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_route_table." + m.generateResourceName(resource) + ".id}",
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapNATGateway maps AWS NAT Gateway resource
func (m *AWSMapper) mapNATGateway(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract subnet ID
	var dependencies []string
	if subnetId, exists := resource.Metadata["subnet_id"]; exists && subnetId != "" {
		config["subnet_id"] = fmt.Sprintf("${aws_subnet.%s.id}", m.cleanSubnetReference(subnetId))
		dependencies = append(dependencies, fmt.Sprintf("aws_subnet.%s", m.cleanSubnetReference(subnetId)))
	}

	// NAT Gateway needs an Elastic IP
	if allocationId, exists := resource.Metadata["allocation_id"]; exists && allocationId != "" {
		config["allocation_id"] = fmt.Sprintf("${aws_eip.%s.id}", m.cleanEIPReference(allocationId))
		dependencies = append(dependencies, fmt.Sprintf("aws_eip.%s", m.cleanEIPReference(allocationId)))
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_nat_gateway." + m.generateResourceName(resource) + ".id}",
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// mapElasticIP maps AWS Elastic IP resource
func (m *AWSMapper) mapElasticIP(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"domain": "vpc", // Assume VPC for modern deployments
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// If associated with an instance
	if instanceId, exists := resource.Metadata["instance_id"]; exists && instanceId != "" {
		config["instance"] = fmt.Sprintf("${aws_instance.%s.id}", m.cleanInstanceReference(instanceId))
	}

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)):         "${aws_eip." + m.generateResourceName(resource) + ".id}",
		fmt.Sprintf("%s_public_ip", m.cleanTerraformName(resource.Name)): "${aws_eip." + m.generateResourceName(resource) + ".public_ip}",
	}

	return config, map[string]generation.Variable{}, outputs, []string{}, nil
}

// mapNetworkACL maps AWS Network ACL resource
func (m *AWSMapper) mapNetworkACL(resource discovery.Resource) (map[string]interface{}, map[string]generation.Variable, map[string]string, []string, error) {
	config := map[string]interface{}{
		"tags": m.mergeTags(resource.Tags, map[string]string{
			"Name": resource.Name,
			"OriginalId": resource.ID,
			"ManagedBy": "chimera",
		}),
	}

	// Extract VPC ID
	var dependencies []string
	if vpcId, exists := resource.Metadata["vpc_id"]; exists && vpcId != "" {
		config["vpc_id"] = fmt.Sprintf("${aws_vpc.%s.id}", m.cleanVPCReference(vpcId))
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.cleanVPCReference(vpcId)))
	}

	// Note: ACL rules would need to be discovered and configured separately

	// Generate outputs
	outputs := map[string]string{
		fmt.Sprintf("%s_id", m.cleanTerraformName(resource.Name)): "${aws_network_acl." + m.generateResourceName(resource) + ".id}",
	}

	return config, map[string]generation.Variable{}, outputs, dependencies, nil
}

// Helper functions for cleaning references
func (m *AWSMapper) cleanVPCReference(vpcId interface{}) string {
	if str, ok := vpcId.(string); ok {
		return m.cleanTerraformName(str)
	}
	return "unknown_vpc"
}

func (m *AWSMapper) cleanSubnetReference(subnetId interface{}) string {
	if str, ok := subnetId.(string); ok {
		return m.cleanTerraformName(str)
	}
	return "unknown_subnet"
}

func (m *AWSMapper) cleanInstanceReference(instanceId interface{}) string {
	if str, ok := instanceId.(string); ok {
		return m.cleanTerraformName(str)
	}
	return "unknown_instance"
}

func (m *AWSMapper) cleanEIPReference(allocationId interface{}) string {
	if str, ok := allocationId.(string); ok {
		return m.cleanTerraformName(str)
	}
	return "unknown_eip"
}

// mergeTags merges discovered tags with generated tags
func (m *AWSMapper) mergeTags(discoveredTags, generatedTags map[string]string) map[string]string {
	merged := make(map[string]string)
	
	// Start with discovered tags
	for k, v := range discoveredTags {
		merged[k] = v
	}
	
	// Add generated tags (these take precedence)
	for k, v := range generatedTags {
		merged[k] = v
	}
	
	return merged
}

// GetProviderConfig returns the AWS provider configuration
func (m *AWSMapper) GetProviderConfig(resources []discovery.Resource) (*generation.ProviderConfig, error) {
	// Determine region from resources
	region := "us-east-1" // default
	if len(resources) > 0 && resources[0].Region != "" {
		region = resources[0].Region
	}

	config := &generation.ProviderConfig{
		Name:     "aws",
		Source:   "hashicorp/aws",
		Version:  "~> 5.0",
		Required: true,
		Config: map[string]interface{}{
			"region": region,
		},
	}

	return config, nil
}

// GetDependencies analyzes resource dependencies
func (m *AWSMapper) GetDependencies(resource discovery.Resource, allResources []discovery.Resource) ([]string, error) {
	var dependencies []string

	switch resource.Type {
	case "aws_subnet":
		// Subnets depend on VPCs
		if vpcId, exists := resource.Metadata["vpc_id"]; exists {
			for _, r := range allResources {
				if r.ID == vpcId && r.Type == "aws_vpc" {
					dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.generateResourceName(r)))
					break
				}
			}
		}
	case "aws_instance":
		// Instances depend on subnets and security groups
		if subnetId, exists := resource.Metadata["subnet_id"]; exists {
			for _, r := range allResources {
				if r.ID == subnetId && r.Type == "aws_subnet" {
					dependencies = append(dependencies, fmt.Sprintf("aws_subnet.%s", m.generateResourceName(r)))
					break
				}
			}
		}
	case "aws_nat_gateway":
		// NAT Gateways depend on subnets and EIPs
		if subnetId, exists := resource.Metadata["subnet_id"]; exists {
			for _, r := range allResources {
				if r.ID == subnetId && r.Type == "aws_subnet" {
					dependencies = append(dependencies, fmt.Sprintf("aws_subnet.%s", m.generateResourceName(r)))
					break
				}
			}
		}
	}

	return dependencies, nil
}

// ValidateMapping validates that the resource mapping is correct
func (m *AWSMapper) ValidateMapping(original discovery.Resource, mapped generation.TerraformResource) error {
	// Basic validation
	if original.Provider != mapped.Provider {
		return fmt.Errorf("provider mismatch: %s != %s", original.Provider, mapped.Provider)
	}

	if original.ID != mapped.SourceInfo.OriginalID {
		return fmt.Errorf("ID mismatch: %s != %s", original.ID, mapped.SourceInfo.OriginalID)
	}

	// Resource-specific validation
	switch mapped.Type {
	case "aws_vpc":
		if _, exists := mapped.Config["cidr_block"]; !exists {
			return fmt.Errorf("VPC missing required cidr_block")
		}
	case "aws_subnet":
		if _, exists := mapped.Config["vpc_id"]; !exists {
			return fmt.Errorf("subnet missing required vpc_id")
		}
		if _, exists := mapped.Config["cidr_block"]; !exists {
			return fmt.Errorf("subnet missing required cidr_block")
		}
	case "aws_instance":
		if _, exists := mapped.Config["instance_type"]; !exists {
			return fmt.Errorf("instance missing required instance_type")
		}
		if _, exists := mapped.Config["ami"]; !exists {
			return fmt.Errorf("instance missing required ami")
		}
	}

	return nil
}
