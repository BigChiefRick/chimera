package providers

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

// GCPConnector implements ProviderConnector for Google Cloud Platform
type GCPConnector struct {
	projectID string
	logger    *logrus.Logger
	clients   map[string]interface{}
}

// GCPConfig contains GCP-specific configuration
type GCPConfig struct {
	ProjectID string   `yaml:"project_id" json:"project_id"`
	Regions   []string `yaml:"regions" json:"regions"`
	Zones     []string `yaml:"zones" json:"zones"`
}

// NewGCPConnector creates a new GCP connector
func NewGCPConnector(ctx context.Context, projectID string) (*GCPConnector, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required for GCP connector")
	}

	connector := &GCPConnector{
		projectID: projectID,
		logger:    logrus.New(),
		clients:   make(map[string]interface{}),
	}

	// Initialize GCP clients
	if err := connector.initializeClients(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize GCP clients: %w", err)
	}

	return connector, nil
}

// initializeClients initializes GCP service clients
func (c *GCPConnector) initializeClients(ctx context.Context) error {
	// Networks client
	networksClient, err := compute.NewNetworksRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create networks client: %w", err)
	}
	c.clients["networks"] = networksClient

	// Subnetworks client
	subnetworksClient, err := compute.NewSubnetworksRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create subnetworks client: %w", err)
	}
	c.clients["subnetworks"] = subnetworksClient

	// Firewalls client
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create firewalls client: %w", err)
	}
	c.clients["firewalls"] = firewallsClient

	// Instances client
	instancesClient, err := compute.NewInstancesRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create instances client: %w", err)
	}
	c.clients["instances"] = instancesClient

	// Zones client
	zonesClient, err := compute.NewZonesRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create zones client: %w", err)
	}
	c.clients["zones"] = zonesClient

	// Regions client
	regionsClient, err := compute.NewRegionsRESTClient(ctx, option.WithUserAgent("chimera-discovery/1.0"))
	if err != nil {
		return fmt.Errorf("failed to create regions client: %w", err)
	}
	c.clients["regions"] = regionsClient

	return nil
}

// Provider returns the cloud provider this connector supports
func (c *GCPConnector) Provider() discovery.CloudProvider {
	return discovery.GCP
}

// ValidateCredentials validates GCP credentials and project access
func (c *GCPConnector) ValidateCredentials(ctx context.Context) error {
	// Test credentials by trying to list regions
	client := c.clients["regions"].(*compute.RegionsClient)
	
	req := &computepb.ListRegionsRequest{
		Project:    c.projectID,
		MaxResults: func() *uint32 { v := uint32(1); return &v }(),
	}
	
	iter := client.List(ctx, req)
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		return fmt.Errorf("GCP credential validation failed: %w", err)
	}

	c.logger.Info("GCP credentials validated successfully")
	return nil
}

// GetRegions returns available GCP regions
func (c *GCPConnector) GetRegions(ctx context.Context) ([]string, error) {
	client := c.clients["regions"].(*compute.RegionsClient)
	
	req := &computepb.ListRegionsRequest{
		Project: c.projectID,
	}
	
	var regions []string
	iter := client.List(ctx, req)
	
	for {
		region, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCP regions: %w", err)
		}
		
		if region.Name != nil {
			regions = append(regions, *region.Name)
		}
	}

	return regions, nil
}

// GetResourceTypes returns available GCP resource types
func (c *GCPConnector) GetResourceTypes(ctx context.Context) ([]string, error) {
	return []string{
		"network",
		"subnetwork",
		"firewall",
		"instance",
	}, nil
}

// Discover discovers GCP resources
func (c *GCPConnector) Discover(ctx context.Context, opts discovery.ProviderDiscoveryOptions) ([]discovery.Resource, error) {
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

	c.logger.Infof("Discovering GCP resources in project: %s", c.projectID)

	// Discover resources for each type
	for _, resourceType := range resourceTypes {
		c.logger.Debugf("Discovering %s resources", resourceType)
		
		typeResources, err := c.discoverResourceType(ctx, resourceType, regions)
		if err != nil {
			c.logger.Warnf("Failed to discover %s resources: %v", resourceType, err)
			continue
		}

		allResources = append(allResources, typeResources...)
	}

	return allResources, nil
}

// discoverResourceType discovers a specific type of GCP resource
func (c *GCPConnector) discoverResourceType(ctx context.Context, resourceType string, regions []string) ([]discovery.Resource, error) {
	switch resourceType {
	case "network":
		return c.discoverNetworks(ctx)
	case "subnetwork":
		return c.discoverSubnetworks(ctx, regions)
	case "firewall":
		return c.discoverFirewalls(ctx)
	case "instance":
		return c.discoverInstances(ctx, regions)
	default:
		c.logger.Warnf("Unsupported resource type: %s", resourceType)
		return nil, nil
	}
}

// discoverNetworks discovers GCP VPC Networks
func (c *GCPConnector) discoverNetworks(ctx context.Context) ([]discovery.Resource, error) {
	client := c.clients["networks"].(*compute.NetworksClient)
	
	req := &computepb.ListNetworksRequest{
		Project: c.projectID,
	}
	
	var resources []discovery.Resource
	iter := client.List(ctx, req)

	for {
		network, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list networks: %w", err)
		}

		resource := discovery.Resource{
			ID:       fmt.Sprintf("projects/%s/global/networks/%s", c.projectID, *network.Name),
			Name:     *network.Name,
			Type:     "gcp_compute_network",
			Provider: discovery.GCP,
			Project:  c.projectID,
			Metadata: make(map[string]interface{}),
			Tags:     make(map[string]string), // GCP uses labels, not tags
		}

		if network.Description != nil {
			resource.Metadata["description"] = *network.Description
		}

		if network.AutoCreateSubnetworks != nil {
			resource.Metadata["auto_create_subnetworks"] = *network.AutoCreateSubnetworks
		}

		if network.RoutingConfig != nil && network.RoutingConfig.RoutingMode != nil {
			resource.Metadata["routing_mode"] = *network.RoutingConfig.RoutingMode
		}

		if network.IPv4Range != nil {
			resource.Metadata["ipv4_range"] = *network.IPv4Range
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSubnetworks discovers GCP Subnetworks
func (c *GCPConnector) discoverSubnetworks(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	client := c.clients["subnetworks"].(*compute.SubnetworksClient)
	
	var resources []discovery.Resource

	// If no regions specified, get all regions
	if len(regions) == 0 {
		var err error
		regions, err = c.GetRegions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions for subnetwork discovery: %w", err)
		}
	}

	for _, region := range regions {
		req := &computepb.ListSubnetworksRequest{
			Project: c.projectID,
			Region:  region,
		}
		
		iter := client.List(ctx, req)

		for {
			subnetwork, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				c.logger.Warnf("Failed to list subnetworks in region %s: %v", region, err)
				break
			}

			resource := discovery.Resource{
				ID:       fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", c.projectID, region, *subnetwork.Name),
				Name:     *subnetwork.Name,
				Type:     "gcp_compute_subnetwork",
				Provider: discovery.GCP,
				Region:   region,
				Project:  c.projectID,
				Metadata: make(map[string]interface{}),
				Tags:     make(map[string]string),
			}

			if subnetwork.Description != nil {
				resource.Metadata["description"] = *subnetwork.Description
			}

			if subnetwork.IpCidrRange != nil {
				resource.Metadata["ip_cidr_range"] = *subnetwork.IpCidrRange
			}

			if subnetwork.Network != nil {
				resource.Metadata["network"] = c.extractNetworkName(*subnetwork.Network)
			}

			if subnetwork.PrivateIpGoogleAccess != nil {
				resource.Metadata["private_ip_google_access"] = *subnetwork.PrivateIpGoogleAccess
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverFirewalls discovers GCP Firewall Rules
func (c *GCPConnector) discoverFirewalls(ctx context.Context) ([]discovery.Resource, error) {
	client := c.clients["firewalls"].(*compute.FirewallsClient)
	
	req := &computepb.ListFirewallsRequest{
		Project: c.projectID,
	}
	
	var resources []discovery.Resource
	iter := client.List(ctx, req)

	for {
		firewall, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list firewalls: %w", err)
		}

		resource := discovery.Resource{
			ID:       fmt.Sprintf("projects/%s/global/firewalls/%s", c.projectID, *firewall.Name),
			Name:     *firewall.Name,
			Type:     "gcp_compute_firewall",
			Provider: discovery.GCP,
			Project:  c.projectID,
			Metadata: make(map[string]interface{}),
			Tags:     make(map[string]string),
		}

		if firewall.Description != nil {
			resource.Metadata["description"] = *firewall.Description
		}

		if firewall.Direction != nil {
			resource.Metadata["direction"] = *firewall.Direction
		}

		if firewall.Priority != nil {
			resource.Metadata["priority"] = *firewall.Priority
		}

		if firewall.Network != nil {
			resource.Metadata["network"] = c.extractNetworkName(*firewall.Network)
		}

		if firewall.Allowed != nil {
			resource.Metadata["allowed_rules_count"] = len(firewall.Allowed)
		}

		if firewall.Denied != nil {
			resource.Metadata["denied_rules_count"] = len(firewall.Denied)
		}

		if firewall.SourceRanges != nil {
			resource.Metadata["source_ranges"] = firewall.SourceRanges
		}

		if firewall.TargetTags != nil {
			resource.Metadata["target_tags"] = firewall.TargetTags
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverInstances discovers GCP Compute Instances
func (c *GCPConnector) discoverInstances(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	// First get all zones for the specified regions
	zones, err := c.getZonesForRegions(ctx, regions)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	client := c.clients["instances"].(*compute.InstancesClient)
	var resources []discovery.Resource

	for _, zone := range zones {
		req := &computepb.ListInstancesRequest{
			Project: c.projectID,
			Zone:    zone.Name,
		}
		
		iter := client.List(ctx, req)

		for {
			instance, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				c.logger.Warnf("Failed to list instances in zone %s: %v", zone.Name, err)
				break
			}

			resource := discovery.Resource{
				ID:       fmt.Sprintf("projects/%s/zones/%s/instances/%s", c.projectID, zone.Name, *instance.Name),
				Name:     *instance.Name,
				Type:     "gcp_compute_instance",
				Provider: discovery.GCP,
				Region:   zone.Region,
				Zone:     zone.Name,
				Project:  c.projectID,
				Metadata: make(map[string]interface{}),
				Tags:     make(map[string]string),
			}

			if instance.Description != nil {
				resource.Metadata["description"] = *instance.Description
			}

			if instance.MachineType != nil {
				resource.Metadata["machine_type"] = c.extractMachineTypeName(*instance.MachineType)
			}

			if instance.Status != nil {
				resource.Metadata["status"] = *instance.Status
			}

			if instance.CreationTimestamp != nil {
				resource.Metadata["creation_timestamp"] = *instance.CreationTimestamp
			}

			// Extract network information
			if instance.NetworkInterfaces != nil && len(instance.NetworkInterfaces) > 0 {
				networkInterface := instance.NetworkInterfaces[0]
				if networkInterface.Network != nil {
					resource.Metadata["network"] = c.extractNetworkName(*networkInterface.Network)
				}
				if networkInterface.Subnetwork != nil {
					resource.Metadata["subnetwork"] = c.extractSubnetworkName(*networkInterface.Subnetwork)
				}
				if networkInterface.NetworkIP != nil {
					resource.Metadata["internal_ip"] = *networkInterface.NetworkIP
				}
				
				// Check for external IP
				if networkInterface.AccessConfigs != nil && len(networkInterface.AccessConfigs) > 0 {
					accessConfig := networkInterface.AccessConfigs[0]
					if accessConfig.NatIP != nil {
						resource.Metadata["external_ip"] = *accessConfig.NatIP
					}
				}
			}

			// Extract boot disk information
			if instance.Disks != nil {
				for _, disk := range instance.Disks {
					if disk.Boot != nil && *disk.Boot {
						if disk.Source != nil {
							resource.Metadata["boot_disk"] = c.extractDiskName(*disk.Source)
						}
						break
					}
				}
			}

			// Convert labels to tags
			if instance.Labels != nil {
				for k, v := range instance.Labels {
					resource.Tags[k] = v
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// Helper functions

// ZoneInfo represents zone information with region mapping
type ZoneInfo struct {
	Name   string
	Region string
}

// getZonesForRegions gets all zones for the specified regions
func (c *GCPConnector) getZonesForRegions(ctx context.Context, regions []string) ([]ZoneInfo, error) {
	client := c.clients["zones"].(*compute.ZonesClient)
	
	req := &computepb.ListZonesRequest{
		Project: c.projectID,
	}
	
	var zones []ZoneInfo
	iter := client.List(ctx, req)

	for {
		zone, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list zones: %w", err)
		}

		if zone.Name == nil || zone.Region == nil {
			continue
		}

		regionName := c.extractRegionFromURL(*zone.Region)
		
		// Filter by regions if specified
		if len(regions) > 0 && !c.containsRegion(regions, regionName) {
			continue
		}

		zones = append(zones, ZoneInfo{
			Name:   *zone.Name,
			Region: regionName,
		})
	}

	return zones, nil
}

// extractNetworkName extracts the network name from a GCP network URL
func (c *GCPConnector) extractNetworkName(networkURL string) string {
	parts := strings.Split(networkURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return networkURL
}

// extractSubnetworkName extracts the subnetwork name from a GCP subnetwork URL
func (c *GCPConnector) extractSubnetworkName(subnetworkURL string) string {
	parts := strings.Split(subnetworkURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return subnetworkURL
}

// extractMachineTypeName extracts the machine type name from a GCP machine type URL
func (c *GCPConnector) extractMachineTypeName(machineTypeURL string) string {
	parts := strings.Split(machineTypeURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return machineTypeURL
}

// extractDiskName extracts the disk name from a GCP disk URL
func (c *GCPConnector) extractDiskName(diskURL string) string {
	parts := strings.Split(diskURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return diskURL
}

// extractRegionFromURL extracts the region name from a GCP region URL
func (c *GCPConnector) extractRegionFromURL(regionURL string) string {
	parts := strings.Split(regionURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return regionURL
}

// containsRegion checks if a region is in the specified regions list
func (c *GCPConnector) containsRegion(regions []string, region string) bool {
	for _, r := range regions {
		if strings.EqualFold(r, region) {
			return true
		}
	}
	return false
}
