package mappers

import (
	"fmt"
	"strings"

	"github.com/BigChiefRick/chimera/pkg/discovery"
	"github.com/BigChiefRick/chimera/pkg/generation"
)

// AWSMapper implements ResourceMapper for AWS resources
type AWSMapper struct{}

// NewAWSMapper creates a new AWS resource mapper
func NewAWSMapper() *AWSMapper {
	return &AWSMapper{}
}

// MapResources maps AWS discovery resources to IaC representations
func (m *AWSMapper) MapResources(resources []discovery.Resource, opts generation.GenerationOptions) ([]generation.MappedResource, error) {
	var mapped []generation.MappedResource

	for _, resource := range resources {
		mappedResource, err := m.mapResource(resource, opts)
		if err != nil {
			// Log warning but continue with other resources
			continue
		}
		if mappedResource != nil {
			mapped = append(mapped, *mappedResource)
		}
	}

	return mapped, nil
}

// mapResource maps a single AWS resource
func (m *AWSMapper) mapResource(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	switch resource.Type {
	case "aws_vpc":
		return m.mapVPC(resource, opts)
	case "aws_subnet":
		return m.mapSubnet(resource, opts)
	case "aws_security_group":
		return m.mapSecurityGroup(resource, opts)
	case "aws_instance":
		return m.mapInstance(resource, opts)
	case "aws_internet_gateway":
		return m.mapInternetGateway(resource, opts)
	case "aws_route_table":
		return m.mapRouteTable(resource, opts)
	case "aws_key_pair":
		return m.mapKeyPair(resource, opts)
	case "aws_ebs_volume":
		return m.mapEBSVolume(resource, opts)
	case "aws_elastic_ip":
		return m.mapElasticIP(resource, opts)
	default:
		return nil, fmt.Errorf("unsupported AWS resource type: %s", resource.Type)
	}
}

// mapVPC maps an AWS VPC to Terraform resource
func (m *AWSMapper) mapVPC(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	// Extract VPC-specific attributes
	cidrBlock := m.getStringFromMetadata(resource.Metadata, "cidr_block", "10.0.0.0/16")
	enableDnsHostnames := m.getBoolFromMetadata(resource.Metadata, "enable_dns_hostnames", true)
	enableDnsSupport := m.getBoolFromMetadata(resource.Metadata, "enable_dns_support", true)

	config := map[string]interface{}{
		"cidr_block":           cidrBlock,
		"enable_dns_hostnames": enableDnsHostnames,
		"enable_dns_support":   enableDnsSupport,
		"tags":                 m.convertTags(resource.Tags),
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_vpc",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     []string{}, // VPCs have no dependencies
		Variables:        m.generateVPCVariables(resource),
		Outputs:          m.generateVPCOutputs(resource),
	}

	return mapped, nil
}

// mapSubnet maps an AWS Subnet to Terraform resource
func (m *AWSMapper) mapSubnet(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	cidrBlock := m.getStringFromMetadata(resource.Metadata, "cidr_block", "10.0.1.0/24")
	vpcId := m.getStringFromMetadata(resource.Metadata, "vpc_id", "")
	availabilityZone := resource.Zone
	mapPublicIpOnLaunch := m.getBoolFromMetadata(resource.Metadata, "map_public_ip_on_launch", false)

	config := map[string]interface{}{
		"vpc_id":                   m.generateVPCReference(vpcId),
		"cidr_block":               cidrBlock,
		"availability_zone":        availabilityZone,
		"map_public_ip_on_launch": mapPublicIpOnLaunch,
		"tags":                     m.convertTags(resource.Tags),
	}

	dependencies := []string{}
	if vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.sanitizeResourceName(vpcId)))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_subnet",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     dependencies,
		Variables:        m.generateSubnetVariables(resource),
		Outputs:          m.generateSubnetOutputs(resource),
	}

	return mapped, nil
}

// mapSecurityGroup maps an AWS Security Group to Terraform resource
func (m *AWSMapper) mapSecurityGroup(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	name := resource.Name
	if name == "" {
		name = m.getStringFromMetadata(resource.Metadata, "group_name", "default")
	}
	
	description := m.getStringFromMetadata(resource.Metadata, "description", "Security group")
	vpcId := m.getStringFromMetadata(resource.Metadata, "vpc_id", "")

	config := map[string]interface{}{
		"name":        name,
		"description": description,
		"vpc_id":      m.generateVPCReference(vpcId),
		"tags":        m.convertTags(resource.Tags),
	}

	// Add ingress and egress rules if available
	if ingressRules := m.getIntFromMetadata(resource.Metadata, "ingress_rules", 0); ingressRules > 0 {
		config["ingress"] = []map[string]interface{}{
			{
				"from_port":   80,
				"to_port":     80,
				"protocol":    "tcp",
				"cidr_blocks": []string{"0.0.0.0/0"},
			},
		}
	}

	if egressRules := m.getIntFromMetadata(resource.Metadata, "egress_rules", 0); egressRules > 0 {
		config["egress"] = []map[string]interface{}{
			{
				"from_port":   0,
				"to_port":     0,
				"protocol":    "-1",
				"cidr_blocks": []string{"0.0.0.0/0"},
			},
		}
	}

	dependencies := []string{}
	if vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.sanitizeResourceName(vpcId)))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_security_group",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     dependencies,
		Variables:        m.generateSecurityGroupVariables(resource),
		Outputs:          m.generateSecurityGroupOutputs(resource),
	}

	return mapped, nil
}

// mapInstance maps an AWS EC2 Instance to Terraform resource
func (m *AWSMapper) mapInstance(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	instanceType := m.getStringFromMetadata(resource.Metadata, "instance_type", "t3.micro")
	imageId := m.getStringFromMetadata(resource.Metadata, "image_id", "ami-0abcdef1234567890")
	subnetId := m.getStringFromMetadata(resource.Metadata, "subnet_id", "")
	vpcId := m.getStringFromMetadata(resource.Metadata, "vpc_id", "")
	keyName := m.getStringFromMetadata(resource.Metadata, "key_name", "")

	config := map[string]interface{}{
		"ami":           imageId,
		"instance_type": instanceType,
		"subnet_id":     m.generateSubnetReference(subnetId),
		"tags":          m.convertTags(resource.Tags),
	}

	if keyName != "" {
		config["key_name"] = keyName
	}

	// Add security groups if available
	config["vpc_security_group_ids"] = []string{"${aws_security_group.default.id}"}

	dependencies := []string{}
	if subnetId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_subnet.%s", m.sanitizeResourceName(subnetId)))
	}
	if vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.sanitizeResourceName(vpcId)))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_instance",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     dependencies,
		Variables:        m.generateInstanceVariables(resource),
		Outputs:          m.generateInstanceOutputs(resource),
	}

	return mapped, nil
}

// mapInternetGateway maps an AWS Internet Gateway to Terraform resource
func (m *AWSMapper) mapInternetGateway(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	vpcId := m.getStringFromMetadata(resource.Metadata, "vpc_id", "")

	config := map[string]interface{}{
		"vpc_id": m.generateVPCReference(vpcId),
		"tags":   m.convertTags(resource.Tags),
	}

	dependencies := []string{}
	if vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.sanitizeResourceName(vpcId)))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_internet_gateway",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     dependencies,
		Variables:        make(map[string]generation.Variable),
		Outputs:          m.generateInternetGatewayOutputs(resource),
	}

	return mapped, nil
}

// mapRouteTable maps an AWS Route Table to Terraform resource
func (m *AWSMapper) mapRouteTable(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	vpcId := m.getStringFromMetadata(resource.Metadata, "vpc_id", "")

	config := map[string]interface{}{
		"vpc_id": m.generateVPCReference(vpcId),
		"tags":   m.convertTags(resource.Tags),
	}

	// Add default route if this is a public route table
	if strings.Contains(resource.Name, "public") || strings.Contains(strings.ToLower(resource.Name), "public") {
		config["route"] = []map[string]interface{}{
			{
				"cidr_block": "0.0.0.0/0",
				"gateway_id": "${aws_internet_gateway.main.id}",
			},
		}
	}

	dependencies := []string{}
	if vpcId != "" {
		dependencies = append(dependencies, fmt.Sprintf("aws_vpc.%s", m.sanitizeResourceName(vpcId)))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_route_table",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     dependencies,
		Variables:        make(map[string]generation.Variable),
		Outputs:          m.generateRouteTableOutputs(resource),
	}

	return mapped, nil
}

// mapKeyPair maps an AWS Key Pair to Terraform resource
func (m *AWSMapper) mapKeyPair(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	keyName := resource.Name
	if keyName == "" {
		keyName = m.getStringFromMetadata(resource.Metadata, "key_name", "default-key")
	}

	config := map[string]interface{}{
		"key_name":   keyName,
		"public_key": "${var.public_key}",
		"tags":       m.convertTags(resource.Tags),
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_key_pair",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     []string{},
		Variables:        m.generateKeyPairVariables(resource),
		Outputs:          m.generateKeyPairOutputs(resource),
	}

	return mapped, nil
}

// mapEBSVolume maps an AWS EBS Volume to Terraform resource
func (m *AWSMapper) mapEBSVolume(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	size := m.getIntFromMetadata(resource.Metadata, "size", 20)
	volumeType := m.getStringFromMetadata(resource.Metadata, "volume_type", "gp3")

	config := map[string]interface{}{
		"availability_zone": resource.Zone,
		"size":              size,
		"type":              volumeType,
		"tags":              m.convertTags(resource.Tags),
	}

	// Add encryption if specified
	if encrypted := m.getBoolFromMetadata(resource.Metadata, "encrypted", false); encrypted {
		config["encrypted"] = true
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_ebs_volume",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     []string{},
		Variables:        m.generateEBSVolumeVariables(resource),
		Outputs:          m.generateEBSVolumeOutputs(resource),
	}

	return mapped, nil
}

// mapElasticIP maps an AWS Elastic IP to Terraform resource
func (m *AWSMapper) mapElasticIP(resource discovery.Resource, opts generation.GenerationOptions) (*generation.MappedResource, error) {
	domain := m.getStringFromMetadata(resource.Metadata, "domain", "vpc")

	config := map[string]interface{}{
		"domain": domain,
		"tags":   m.convertTags(resource.Tags),
	}

	// Associate with instance if specified
	if instanceId := m.getStringFromMetadata(resource.Metadata, "instance_id", ""); instanceId != "" {
		config["instance"] = fmt.Sprintf("${aws_instance.%s.id}", m.sanitizeResourceName(instanceId))
	}

	mapped := &generation.MappedResource{
		OriginalResource: resource,
		ResourceType:     "aws_eip",
		ResourceName:     m.generateResourceName(resource),
		Configuration:    config,
		Dependencies:     []string{},
		Variables:        make(map[string]generation.Variable),
		Outputs:          m.generateElasticIPOutputs(resource),
	}

	return mapped, nil
}

// Helper methods

// generateResourceName creates a Terraform-safe resource name
func (m *AWSMapper) generateResourceName(resource discovery.Resource) string {
	name := resource.Name
	if name == "" {
		// Extract name from ID
		parts := strings.Split(resource.ID, "-")
		if len(parts) > 1 {
			name = parts[len(parts)-1]
		} else {
			name = resource.ID
		}
	}
	return m.sanitizeResourceName(name)
}

// sanitizeResourceName sanitizes a string for use as Terraform resource name
func (m *AWSMapper) sanitizeResourceName(name string) string {
	// Replace invalid characters with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	
	// Remove any remaining non-alphanumeric characters except underscores
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	
	cleaned := result.String()
	
	// Ensure it starts with a letter or underscore
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "resource_" + cleaned
	}
	
	if cleaned == "" {
		cleaned = "resource"
	}
	
	return cleaned
}

// generateVPCReference generates a reference to a VPC resource
func (m *AWSMapper) generateVPCReference(vpcId string) string {
	if vpcId == "" {
		return "${aws_vpc.main.id}"
	}
	return fmt.Sprintf("${aws_vpc.%s.id}", m.sanitizeResourceName(vpcId))
}

// generateSubnetReference generates a reference to a subnet resource
func (m *AWSMapper) generateSubnetReference(subnetId string) string {
	if subnetId == "" {
		return "${aws_subnet.main.id}"
	}
	return fmt.Sprintf("${aws_subnet.%s.id}", m.sanitizeResourceName(subnetId))
}

// convertTags converts discovery tags to Terraform format
func (m *AWSMapper) convertTags(tags map[string]string) map[string]string {
	if tags == nil {
		return map[string]string{
			"ManagedBy": "Chimera",
		}
	}
	
	// Add Chimera management tag
	result := make(map[string]string)
	for k, v := range tags {
		result[k] = v
	}
	result["ManagedBy"] = "Chimera"
	
	return result
}

// Metadata helper methods
func (m *AWSMapper) getStringFromMetadata(metadata map[string]interface{}, key, defaultValue string) string {
	if value, exists := metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (m *AWSMapper) getBoolFromMetadata(metadata map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := metadata[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (m *AWSMapper) getIntFromMetadata(metadata map[string]interface{}, key string, defaultValue int) int {
	if value, exists := metadata[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
		if i32, ok := value.(int32); ok {
			return int(i32)
		}
		if i64, ok := value.(int64); ok {
			return int(i64)
		}
	}
	return defaultValue
}

// Variable generation methods
func (m *AWSMapper) generateVPCVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"vpc_cidr": {
			Type:        "string",
			Description: "CIDR block for the VPC",
			Default:     m.getStringFromMetadata(resource.Metadata, "cidr_block", "10.0.0.0/16"),
		},
	}
}

func (m *AWSMapper) generateSubnetVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"subnet_cidr": {
			Type:        "string",
			Description: "CIDR block for the subnet",
			Default:     m.getStringFromMetadata(resource.Metadata, "cidr_block", "10.0.1.0/24"),
		},
	}
}

func (m *AWSMapper) generateSecurityGroupVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"sg_name": {
			Type:        "string",
			Description: "Name for the security group",
			Default:     resource.Name,
		},
	}
}

func (m *AWSMapper) generateInstanceVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"instance_type": {
			Type:        "string",
			Description: "EC2 instance type",
			Default:     m.getStringFromMetadata(resource.Metadata, "instance_type", "t3.micro"),
		},
		"ami_id": {
			Type:        "string",
			Description: "AMI ID for the instance",
			Default:     m.getStringFromMetadata(resource.Metadata, "image_id", "ami-0abcdef1234567890"),
		},
	}
}

func (m *AWSMapper) generateKeyPairVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"public_key": {
			Type:        "string",
			Description: "Public key for the key pair",
		},
	}
}

func (m *AWSMapper) generateEBSVolumeVariables(resource discovery.Resource) map[string]generation.Variable {
	return map[string]generation.Variable{
		"volume_size": {
			Type:        "number",
			Description: "Size of the EBS volume in GB",
			Default:     m.getIntFromMetadata(resource.Metadata, "size", 20),
		},
	}
}

// Output generation methods
func (m *AWSMapper) generateVPCOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"vpc_id": {
			Description: "ID of the VPC",
			Value:       fmt.Sprintf("${aws_vpc.%s.id}", resourceName),
		},
		"vpc_cidr_block": {
			Description: "CIDR block of the VPC",
			Value:       fmt.Sprintf("${aws_vpc.%s.cidr_block}", resourceName),
		},
	}
}

func (m *AWSMapper) generateSubnetOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"subnet_id": {
			Description: "ID of the subnet",
			Value:       fmt.Sprintf("${aws_subnet.%s.id}", resourceName),
		},
	}
}

func (m *AWSMapper) generateSecurityGroupOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"security_group_id": {
			Description: "ID of the security group",
			Value:       fmt.Sprintf("${aws_security_group.%s.id}", resourceName),
		},
	}
}

func (m *AWSMapper) generateInstanceOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"instance_id": {
			Description: "ID of the EC2 instance",
			Value:       fmt.Sprintf("${aws_instance.%s.id}", resourceName),
		},
		"instance_public_ip": {
			Description: "Public IP of the EC2 instance",
			Value:       fmt.Sprintf("${aws_instance.%s.public_ip}", resourceName),
		},
		"instance_private_ip": {
			Description: "Private IP of the EC2 instance",
			Value:       fmt.Sprintf("${aws_instance.%s.private_ip}", resourceName),
		},
	}
}

func (m *AWSMapper) generateInternetGatewayOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"internet_gateway_id": {
			Description: "ID of the Internet Gateway",
			Value:       fmt.Sprintf("${aws_internet_gateway.%s.id}", resourceName),
		},
	}
}

func (m *AWSMapper) generateRouteTableOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"route_table_id": {
			Description: "ID of the route table",
			Value:       fmt.Sprintf("${aws_route_table.%s.id}", resourceName),
		},
	}
}

func (m *AWSMapper) generateKeyPairOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"key_pair_name": {
			Description: "Name of the key pair",
			Value:       fmt.Sprintf("${aws_key_pair.%s.key_name}", resourceName),
		},
	}
}

func (m *AWSMapper) generateEBSVolumeOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"volume_id": {
			Description: "ID of the EBS volume",
			Value:       fmt.Sprintf("${aws_ebs_volume.%s.id}", resourceName),
		},
	}
}

func (m *AWSMapper) generateElasticIPOutputs(resource discovery.Resource) map[string]generation.Output {
	resourceName := m.generateResourceName(resource)
	return map[string]generation.Output{
		"elastic_ip": {
			Description: "Elastic IP address",
			Value:       fmt.Sprintf("${aws_eip.%s.public_ip}", resourceName),
		},
	}
}
