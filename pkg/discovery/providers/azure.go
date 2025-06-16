package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/sirupsen/logrus"

	"github.com/BigChiefRick/chimera/pkg/discovery"
)

// AzureConnector implements ProviderConnector for Azure
type AzureConnector struct {
	credential     azcore.TokenCredential
	subscriptionID string
	logger         *logrus.Logger
	clients        map[string]interface{}
}

// AzureConfig contains Azure-specific configuration
type AzureConfig struct {
	SubscriptionID string   `yaml:"subscription_id" json:"subscription_id"`
	TenantID       string   `yaml:"tenant_id" json:"tenant_id"`
	Locations      []string `yaml:"locations" json:"locations"`
}

// NewAzureConnector creates a new Azure connector
func NewAzureConnector(ctx context.Context, subscriptionID string) (*AzureConnector, error) {
	// Use DefaultAzureCredential which handles various auth methods
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	connector := &AzureConnector{
		credential:     credential,
		subscriptionID: subscriptionID,
		logger:         logrus.New(),
		clients:        make(map[string]interface{}),
	}

	// Initialize Azure clients
	if err := connector.initializeClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize Azure clients: %w", err)
	}

	return connector, nil
}

// initializeClients initializes Azure service clients
func (c *AzureConnector) initializeClients() error {
	clientOptions := &arm.ClientOptions{}

	// Resource Groups client
	resourceGroupsClient, err := armresources.NewResourceGroupsClient(c.subscriptionID, c.credential, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create resource groups client: %w", err)
	}
	c.clients["resourceGroups"] = resourceGroupsClient

	// Virtual Networks client
	virtualNetworksClient, err := armnetwork.NewVirtualNetworksClient(c.subscriptionID, c.credential, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create virtual networks client: %w", err)
	}
	c.clients["virtualNetworks"] = virtualNetworksClient

	// Subnets client
	subnetsClient, err := armnetwork.NewSubnetsClient(c.subscriptionID, c.credential, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create subnets client: %w", err)
	}
	c.clients["subnets"] = subnetsClient

	// Network Security Groups client
	nsgClient, err := armnetwork.NewSecurityGroupsClient(c.subscriptionID, c.credential, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create NSG client: %w", err)
	}
	c.clients["networkSecurityGroups"] = nsgClient

	// Virtual Machines client
	virtualMachinesClient, err := armcompute.NewVirtualMachinesClient(c.subscriptionID, c.credential, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create virtual machines client: %w", err)
	}
	c.clients["virtualMachines"] = virtualMachinesClient

	return nil
}

// Provider returns the cloud provider this connector supports
func (c *AzureConnector) Provider() discovery.CloudProvider {
	return discovery.Azure
}

// ValidateCredentials validates Azure credentials
func (c *AzureConnector) ValidateCredentials(ctx context.Context) error {
	// Test credentials by trying to list resource groups
	client := c.clients["resourceGroups"].(*armresources.ResourceGroupsClient)
	
	pager := client.NewListPager(&armresources.ResourceGroupsClientListOptions{
		Top: func() *int32 { v := int32(1); return &v }(),
	})
	
	_, err := pager.NextPage(ctx)
	if err != nil {
		return fmt.Errorf("Azure credential validation failed: %w", err)
	}

	c.logger.Info("Azure credentials validated successfully")
	return nil
}

// GetRegions returns available Azure regions (locations)
func (c *AzureConnector) GetRegions(ctx context.Context) ([]string, error) {
	// Azure locations are typically known, but we can return common ones
	// In a real implementation, you might want to query the Subscriptions API
	regions := []string{
		"eastus", "eastus2", "westus", "westus2", "westus3",
		"centralus", "northcentralus", "southcentralus", "westcentralus",
		"northeurope", "westeurope", "uksouth", "ukwest",
		"southeastasia", "eastasia", "australiaeast", "australiasoutheast",
		"japaneast", "japanwest", "koreacentral", "koreasouth",
		"canadacentral", "canadaeast", "brazilsouth",
		"southafricanorth", "uaenorth",
	}

	return regions, nil
}

// GetResourceTypes returns available Azure resource types
func (c *AzureConnector) GetResourceTypes(ctx context.Context) ([]string, error) {
	return []string{
		"resource_group",
		"virtual_network",
		"subnet",
		"network_security_group",
		"virtual_machine",
	}, nil
}

// Discover discovers Azure resources
func (c *AzureConnector) Discover(ctx context.Context, opts discovery.ProviderDiscoveryOptions) ([]discovery.Resource, error) {
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

	// Azure resources are subscription-wide, but we'll filter by location
	c.logger.Infof("Discovering Azure resources in subscription: %s", c.subscriptionID)

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

// discoverResourceType discovers a specific type of Azure resource
func (c *AzureConnector) discoverResourceType(ctx context.Context, resourceType string, regions []string) ([]discovery.Resource, error) {
	switch resourceType {
	case "resource_group":
		return c.discoverResourceGroups(ctx, regions)
	case "virtual_network":
		return c.discoverVirtualNetworks(ctx, regions)
	case "subnet":
		return c.discoverSubnets(ctx, regions)
	case "network_security_group":
		return c.discoverNetworkSecurityGroups(ctx, regions)
	case "virtual_machine":
		return c.discoverVirtualMachines(ctx, regions)
	default:
		c.logger.Warnf("Unsupported resource type: %s", resourceType)
		return nil, nil
	}
}

// discoverResourceGroups discovers Azure Resource Groups
func (c *AzureConnector) discoverResourceGroups(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	client := c.clients["resourceGroups"].(*armresources.ResourceGroupsClient)
	
	var resources []discovery.Resource
	pager := client.NewListPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list resource groups: %w", err)
		}

		for _, rg := range page.Value {
			if rg.Name == nil || rg.Location == nil {
				continue
			}

			// Filter by regions if specified
			if len(regions) > 0 && !c.containsLocation(regions, *rg.Location) {
				continue
			}

			resource := discovery.Resource{
				ID:           *rg.ID,
				Name:         *rg.Name,
				Type:         "azure_resource_group",
				Provider:     discovery.Azure,
				Region:       *rg.Location,
				ResourceGroup: *rg.Name,
				Metadata:     make(map[string]interface{}),
				Tags:         c.convertAzureTags(rg.Tags),
			}

			if rg.Properties != nil && rg.Properties.ProvisioningState != nil {
				resource.Metadata["provisioning_state"] = *rg.Properties.ProvisioningState
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverVirtualNetworks discovers Azure Virtual Networks
func (c *AzureConnector) discoverVirtualNetworks(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	client := c.clients["virtualNetworks"].(*armnetwork.VirtualNetworksClient)
	
	var resources []discovery.Resource
	pager := client.NewListAllPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list virtual networks: %w", err)
		}

		for _, vnet := range page.Value {
			if vnet.Name == nil || vnet.Location == nil {
				continue
			}

			// Filter by regions if specified
			if len(regions) > 0 && !c.containsLocation(regions, *vnet.Location) {
				continue
			}

			resource := discovery.Resource{
				ID:           *vnet.ID,
				Name:         *vnet.Name,
				Type:         "azure_virtual_network",
				Provider:     discovery.Azure,
				Region:       *vnet.Location,
				ResourceGroup: c.extractResourceGroupFromID(*vnet.ID),
				Metadata:     make(map[string]interface{}),
				Tags:         c.convertAzureTags(vnet.Tags),
			}

			if vnet.Properties != nil {
				if vnet.Properties.ProvisioningState != nil {
					resource.Metadata["provisioning_state"] = *vnet.Properties.ProvisioningState
				}
				
				if vnet.Properties.AddressSpace != nil && vnet.Properties.AddressSpace.AddressPrefixes != nil {
					resource.Metadata["address_prefixes"] = vnet.Properties.AddressSpace.AddressPrefixes
				}

				if vnet.Properties.Subnets != nil {
					resource.Metadata["subnet_count"] = len(vnet.Properties.Subnets)
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverSubnets discovers Azure Subnets
func (c *AzureConnector) discoverSubnets(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	// First get all virtual networks, then get subnets for each
	vnets, err := c.discoverVirtualNetworks(ctx, regions)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual networks for subnet discovery: %w", err)
	}

	client := c.clients["subnets"].(*armnetwork.SubnetsClient)
	var resources []discovery.Resource

	for _, vnet := range vnets {
		pager := client.NewListPager(vnet.ResourceGroup, vnet.Name, nil)

		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				c.logger.Warnf("Failed to list subnets for VNet %s: %v", vnet.Name, err)
				continue
			}

			for _, subnet := range page.Value {
				if subnet.Name == nil {
					continue
				}

				resource := discovery.Resource{
					ID:           *subnet.ID,
					Name:         *subnet.Name,
					Type:         "azure_subnet",
					Provider:     discovery.Azure,
					Region:       vnet.Region,
					ResourceGroup: vnet.ResourceGroup,
					Metadata:     make(map[string]interface{}),
					Tags:         make(map[string]string), // Subnets don't have tags in Azure
				}

				resource.Metadata["virtual_network"] = vnet.Name

				if subnet.Properties != nil {
					if subnet.Properties.ProvisioningState != nil {
						resource.Metadata["provisioning_state"] = *subnet.Properties.ProvisioningState
					}
					
					if subnet.Properties.AddressPrefix != nil {
						resource.Metadata["address_prefix"] = *subnet.Properties.AddressPrefix
					}
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// discoverNetworkSecurityGroups discovers Azure Network Security Groups
func (c *AzureConnector) discoverNetworkSecurityGroups(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	client := c.clients["networkSecurityGroups"].(*armnetwork.SecurityGroupsClient)
	
	var resources []discovery.Resource
	pager := client.NewListAllPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list network security groups: %w", err)
		}

		for _, nsg := range page.Value {
			if nsg.Name == nil || nsg.Location == nil {
				continue
			}

			// Filter by regions if specified
			if len(regions) > 0 && !c.containsLocation(regions, *nsg.Location) {
				continue
			}

			resource := discovery.Resource{
				ID:           *nsg.ID,
				Name:         *nsg.Name,
				Type:         "azure_network_security_group",
				Provider:     discovery.Azure,
				Region:       *nsg.Location,
				ResourceGroup: c.extractResourceGroupFromID(*nsg.ID),
				Metadata:     make(map[string]interface{}),
				Tags:         c.convertAzureTags(nsg.Tags),
			}

			if nsg.Properties != nil {
				if nsg.Properties.ProvisioningState != nil {
					resource.Metadata["provisioning_state"] = *nsg.Properties.ProvisioningState
				}

				if nsg.Properties.SecurityRules != nil {
					resource.Metadata["security_rules_count"] = len(nsg.Properties.SecurityRules)
				}

				if nsg.Properties.DefaultSecurityRules != nil {
					resource.Metadata["default_security_rules_count"] = len(nsg.Properties.DefaultSecurityRules)
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverVirtualMachines discovers Azure Virtual Machines
func (c *AzureConnector) discoverVirtualMachines(ctx context.Context, regions []string) ([]discovery.Resource, error) {
	client := c.clients["virtualMachines"].(*armcompute.VirtualMachinesClient)
	
	var resources []discovery.Resource
	pager := client.NewListAllPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list virtual machines: %w", err)
		}

		for _, vm := range page.Value {
			if vm.Name == nil || vm.Location == nil {
				continue
			}

			// Filter by regions if specified
			if len(regions) > 0 && !c.containsLocation(regions, *vm.Location) {
				continue
			}

			resource := discovery.Resource{
				ID:           *vm.ID,
				Name:         *vm.Name,
				Type:         "azure_virtual_machine",
				Provider:     discovery.Azure,
				Region:       *vm.Location,
				ResourceGroup: c.extractResourceGroupFromID(*vm.ID),
				Metadata:     make(map[string]interface{}),
				Tags:         c.convertAzureTags(vm.Tags),
			}

			if vm.Properties != nil {
				if vm.Properties.ProvisioningState != nil {
					resource.Metadata["provisioning_state"] = *vm.Properties.ProvisioningState
				}

				if vm.Properties.HardwareProfile != nil && vm.Properties.HardwareProfile.VMSize != nil {
					resource.Metadata["vm_size"] = *vm.Properties.HardwareProfile.VMSize
				}

				if vm.Properties.StorageProfile != nil && vm.Properties.StorageProfile.ImageReference != nil {
					imageRef := vm.Properties.StorageProfile.ImageReference
					if imageRef.Publisher != nil {
						resource.Metadata["image_publisher"] = *imageRef.Publisher
					}
					if imageRef.Offer != nil {
						resource.Metadata["image_offer"] = *imageRef.Offer
					}
					if imageRef.SKU != nil {
						resource.Metadata["image_sku"] = *imageRef.SKU
					}
				}

				if vm.Properties.OSProfile != nil {
					if vm.Properties.OSProfile.ComputerName != nil {
						resource.Metadata["computer_name"] = *vm.Properties.OSProfile.ComputerName
					}
					if vm.Properties.OSProfile.AdminUsername != nil {
						resource.Metadata["admin_username"] = *vm.Properties.OSProfile.AdminUsername
					}
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// Helper functions

// extractResourceGroupFromID extracts the resource group name from an Azure resource ID
func (c *AzureConnector) extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// convertAzureTags converts Azure tags to a map
func (c *AzureConnector) convertAzureTags(tags map[string]*string) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if v != nil {
			result[k] = *v
		}
	}
	return result
}

// containsLocation checks if a location is in the specified regions list
func (c *AzureConnector) containsLocation(regions []string, location string) bool {
	for _, region := range regions {
		if strings.EqualFold(region, location) {
			return true
		}
	}
	return false
}
