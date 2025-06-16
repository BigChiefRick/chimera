package providers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

// AWSConnector implements ProviderConnector for AWS
type AWSConnector struct {
	config aws.Config
	logger *logrus.Logger
	clients map[string]interface{}
}

// NewAWSConnector creates a new AWS connector
func NewAWSConnector(ctx context.Context, region string) (*AWSConnector, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	connector := &AWSConnector{
		config:  cfg,
		logger:  logrus.New(),
		clients: make(map[string]interface{}),
	}

	// Initialize only working clients
	connector.clients["ec2"] = ec2.NewFromConfig(cfg)
	connector.clients["sts"] = sts.NewFromConfig(cfg)

	return connector, nil
}

// Provider returns the cloud provider this connector supports
func (c *AWSConnector) Provider() discovery.CloudProvider {
	return discovery.AWS
}

// ValidateCredentials validates AWS credentials
func (c *AWSConnector) ValidateCredentials(ctx context.Context) error {
	stsClient := c.clients["sts"].(*sts.Client)
	
	_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("AWS credential validation failed: %w", err)
	}

	c.logger.Info("AWS credentials validated successfully")
	return nil
}

// GetRegions returns available AWS regions
func (c *AWSConnector) GetRegions(ctx context.Context) ([]string, error) {
	ec2Client := c.clients["ec2"].(*ec2.Client)
	
	result, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS regions: %w", err)
	}

	regions := make([]string, len(result.Regions))
	for i, region := range result.Regions {
		regions[i] = aws.ToString(region.RegionName)
	}

	return regions, nil
}

// GetResourceTypes returns available AWS resource types
func (c *AWSConnector) GetResourceTypes(ctx context.Context) ([]string, error) {
	// Return only supported resource types for Phase 1
	return []string{
		"vpc",
		"subnet",
		"security_group",
		"instance",
	}, nil
}

// Discover discovers AWS resources
func (c *AWSConnector) Discover(ctx context.Context, opts discovery.ProviderDiscoveryOptions) ([]discovery.Resource, error) {
	var allResources []discovery.Resource

	// Get regions to scan
	regions := opts.Regions
	if len(regions) == 0 {
		var err error
		regions, err = c.GetRegions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions: %w", err)
		}
	}

	// Get resource types to discover
	resourceTypes := opts.ResourceTypes
	if len(resourceTypes) == 0 {
		var err error
		resourceTypes, err = c.GetResourceTypes(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource types: %w", err)
		}
	}

	// Discover resources for each region
	for _, region := range regions {
		c.logger.Infof("Discovering AWS resources in region: %s", region)
		
		// Create region-specific connector
		regionConnector, err := NewAWSConnector(ctx, region)
		if err != nil {
			c.logger.Warnf("Failed to create connector for region %s: %v", region, err)
			continue
		}

		regionResources, err := regionConnector.discoverRegionResources(ctx, region, resourceTypes)
		if err != nil {
			c.logger.Warnf("Failed to discover resources in region %s: %v", region, err)
			continue
		}

		allResources = append(allResources, regionResources...)
	}

	return allResources, nil
}

// discoverRegionResources discovers resources in a specific region
func (c *AWSConnector) discoverRegionResources(ctx context.Context, region string, resourceTypes []string) ([]discovery.Resource, error) {
	var resources []discovery.Resource

	for _, resourceType := range resourceTypes {
		c.logger.Debugf("Discovering %s resources in region %s", resourceType, region)
		
		typeResources, err := c.discoverResourceType(ctx, region, resourceType)
		if err != nil {
			c.logger.Warnf("Failed to discover %s resources in region %s: %v", resourceType, region, err)
			continue
		}

		resources = append(resources, typeResources...)
	}

	return resources, nil
}

// discoverResourceType discovers a specific type of AWS resource
func (c *AWSConnector) discoverResourceType(ctx context.Context, region, resourceType string) ([]discovery.Resource, error) {
	switch resourceType {
	case "vpc":
		return c.discoverVPCs(ctx, region)
	case "subnet":
		return c.discoverSubnets(ctx, region)
	case "security_group":
		return c.discoverSecurityGroups(ctx, region)
	case "instance":
		return c.discoverInstances(ctx, region)
	default:
		c.logger.Warnf("Unsupported resource type: %s", resourceType)
		return nil, nil
	}
}

// discoverVPCs discovers VPCs
func (c *AWSConnector) discoverVPCs(ctx context.Context, region string) ([]discovery.Resource, error) {
	ec2Client := c.clients["ec2"].(*ec2.Client)
	
	result, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	var resources []discovery.Resource
	for _, vpc := range result.Vpcs {
		resource := discovery.Resource{
			ID:       aws.ToString(vpc.VpcId),
			Name:     c.getNameFromTags(vpc.Tags),
			Type:     "aws_vpc",
			Provider: discovery.AWS,
			Region:   region,
			Metadata: map[string]interface{}{
				"cidr_block": aws.ToString(vpc.CidrBlock),
				"state":      string(vpc.State),
				"is_default": aws.ToBool(vpc.IsDefault),
			},
			Tags: c.convertAWSTags(vpc.Tags),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSubnets discovers subnets - FIXED VERSION
func (c *AWSConnector) discoverSubnets(ctx context.Context, region string) ([]discovery.Resource, error) {
	ec2Client := c.clients["ec2"].(*ec2.Client)
	
	// FIXED: Corrected the function call syntax
	result, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	var resources []discovery.Resource
	for _, subnet := range result.Subnets {
		resource := discovery.Resource{
			ID:       aws.ToString(subnet.SubnetId),
			Name:     c.getNameFromTags(subnet.Tags),
			Type:     "aws_subnet",
			Provider: discovery.AWS,
			Region:   region,
			Zone:     aws.ToString(subnet.AvailabilityZone),
			Metadata: map[string]interface{}{
				"vpc_id":                        aws.ToString(subnet.VpcId),
				"cidr_block":                    aws.ToString(subnet.CidrBlock),
				"state":                         string(subnet.State),
				"map_public_ip_on_launch":       aws.ToBool(subnet.MapPublicIpOnLaunch),
				"available_ip_address_count":    aws.ToInt32(subnet.AvailableIpAddressCount),
			},
			Tags: c.convertAWSTags(subnet.Tags),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSecurityGroups discovers security groups
func (c *AWSConnector) discoverSecurityGroups(ctx context.Context, region string) ([]discovery.Resource, error) {
	ec2Client := c.clients["ec2"].(*ec2.Client)
	
	result, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %w", err)
	}

	var resources []discovery.Resource
	for _, sg := range result.SecurityGroups {
		resource := discovery.Resource{
			ID:       aws.ToString(sg.GroupId),
			Name:     aws.ToString(sg.GroupName),
			Type:     "aws_security_group",
			Provider: discovery.AWS,
			Region:   region,
			Metadata: map[string]interface{}{
				"vpc_id":      aws.ToString(sg.VpcId),
				"description": aws.ToString(sg.Description),
				"owner_id":    aws.ToString(sg.OwnerId),
			},
			Tags: c.convertAWSTags(sg.Tags),
		}

		// Add rule counts
		resource.Metadata["ingress_rules"] = len(sg.IpPermissions)
		resource.Metadata["egress_rules"] = len(sg.IpPermissionsEgress)

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverInstances discovers EC2 instances
func (c *AWSConnector) discoverInstances(ctx context.Context, region string) ([]discovery.Resource, error) {
	ec2Client := c.clients["ec2"].(*ec2.Client)
	
	result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var resources []discovery.Resource
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			resource := discovery.Resource{
				ID:       aws.ToString(instance.InstanceId),
				Name:     c.getNameFromTags(instance.Tags),
				Type:     "aws_instance",
				Provider: discovery.AWS,
				Region:   region,
				Zone:     aws.ToString(instance.Placement.AvailabilityZone),
				Metadata: map[string]interface{}{
					"instance_type":   string(instance.InstanceType),
					"state":           string(instance.State.Name),
					"image_id":        aws.ToString(instance.ImageId),
					"vpc_id":          aws.ToString(instance.VpcId),
					"subnet_id":       aws.ToString(instance.SubnetId),
					"private_ip":      aws.ToString(instance.PrivateIpAddress),
					"public_ip":       aws.ToString(instance.PublicIpAddress),
				},
				Tags: c.convertAWSTags(instance.Tags),
			}

			if instance.LaunchTime != nil {
				resource.CreatedAt = instance.LaunchTime
			}

			if instance.KeyName != nil {
				resource.Metadata["key_name"] = aws.ToString(instance.KeyName)
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// Helper functions

// getNameFromTags extracts the Name tag from AWS tags
func (c *AWSConnector) getNameFromTags(tags []ec2Types.Tag) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// convertAWSTags converts AWS EC2 tags to a map
func (c *AWSConnector) convertAWSTags(tags []ec2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}