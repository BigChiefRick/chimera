package discovery

import (
	"context"
	"time"
)

// CloudProvider represents the different cloud providers supported
type CloudProvider string

const (
	AWS       CloudProvider = "aws"
	Azure     CloudProvider = "azure"
	GCP       CloudProvider = "gcp"
	VMware    CloudProvider = "vmware"
	KVM       CloudProvider = "kvm"
	OpenStack CloudProvider = "openstack"
)

// Resource represents a discovered cloud resource
type Resource struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Provider    CloudProvider          `json:"provider"`
	Region      string                 `json:"region"`
	Zone        string                 `json:"zone,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	CreatedAt   *time.Time             `json:"created_at,omitempty"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
	Dependencies []string              `json:"dependencies,omitempty"`
}

// DiscoveryOptions contains configuration for resource discovery
type DiscoveryOptions struct {
	Providers        []CloudProvider       `json:"providers"`
	Regions          []string              `json:"regions,omitempty"`
	ResourceTypes    []string              `json:"resource_types,omitempty"`
	Tags             map[string]string     `json:"tags,omitempty"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	IncludeManaged   bool                  `json:"include_managed"`
	IncludeDefaults  bool                  `json:"include_defaults"`
	MaxConcurrency   int                   `json:"max_concurrency,omitempty"`
	Timeout          time.Duration         `json:"timeout,omitempty"`
	UseCache         bool                  `json:"use_cache"`
	CacheTTL         time.Duration         `json:"cache_ttl,omitempty"`
}

// DiscoveryResult contains the results of a discovery operation
type DiscoveryResult struct {
	Resources []Resource        `json:"resources"`
	Errors    []DiscoveryError  `json:"errors,omitempty"`
	Metadata  DiscoveryMetadata `json:"metadata"`
}

// DiscoveryError represents an error during discovery
type DiscoveryError struct {
	Provider     CloudProvider `json:"provider"`
	Region       string        `json:"region,omitempty"`
	ResourceType string        `json:"resource_type,omitempty"`
	Message      string        `json:"message"`
	Error        error         `json:"-"`
	Severity     string        `json:"severity"`
	Timestamp    time.Time     `json:"timestamp"`
}

// DiscoveryMetadata contains metadata about the discovery operation
type DiscoveryMetadata struct {
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	ResourceCount int                    `json:"resource_count"`
	ProviderStats map[string]int         `json:"provider_stats"`
	Filters       map[string]interface{} `json:"filters"`
	ErrorCount    int                    `json:"error_count"`
	WarningCount  int                    `json:"warning_count"`
}

// DiscoveryEngine defines the main interface for infrastructure discovery
type DiscoveryEngine interface {
	// Discover discovers resources based on the provided options
	Discover(ctx context.Context, opts DiscoveryOptions) (*DiscoveryResult, error)
	
	// RegisterConnector registers a provider connector
	RegisterConnector(connector ProviderConnector)
	
	// ListProviders returns the list of supported providers
	ListProviders() []CloudProvider
	
	// ValidateCredentials validates credentials for the specified providers
	ValidateCredentials(ctx context.Context, providers []CloudProvider) error
	
	// GetProviderRegions returns available regions for a provider
	GetProviderRegions(ctx context.Context, provider CloudProvider) ([]string, error)
	
	// GetResourceTypes returns available resource types for a provider
	GetResourceTypes(ctx context.Context, provider CloudProvider) ([]string, error)
}

// ProviderConnector defines the interface for cloud provider connectors
type ProviderConnector interface {
	// Provider returns the cloud provider this connector handles
	Provider() CloudProvider
	
	// Connect establishes connection to the provider
	Connect(ctx context.Context) error
	
	// Disconnect closes the connection to the provider
	Disconnect(ctx context.Context) error
	
	// ValidateCredentials validates the credentials for this provider
	ValidateCredentials(ctx context.Context) error
	
	// DiscoverResources discovers resources for this provider
	DiscoverResources(ctx context.Context, opts ProviderDiscoveryOptions) ([]Resource, error)
	
	// GetRegions returns available regions for this provider
	GetRegions(ctx context.Context) ([]string, error)
	
	// GetResourceTypes returns available resource types for this provider
	GetResourceTypes(ctx context.Context) ([]string, error)
	
	// GetResourcesByType discovers resources of a specific type
	GetResourcesByType(ctx context.Context, resourceType string, region string) ([]Resource, error)
}

// ProviderDiscoveryOptions contains provider-specific discovery options
type ProviderDiscoveryOptions struct {
	Provider      CloudProvider          `json:"provider"`
	Regions       []string               `json:"regions,omitempty"`
	ResourceTypes []string               `json:"resource_types,omitempty"`
	Filters       map[string]interface{} `json:"filters,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	IncludeManaged bool                  `json:"include_managed"`
	IncludeDefaults bool                 `json:"include_defaults"`
}

// SteampipeConnector defines the interface for Steampipe integration
type SteampipeConnector interface {
	// Connect establishes connection to Steampipe
	Connect(ctx context.Context) error
	
	// Disconnect closes the Steampipe connection
	Disconnect(ctx context.Context) error
	
	// IsConnected returns true if connected to Steampipe
	IsConnected() bool
	
	// Query executes a query against Steampipe
	Query(ctx context.Context, query string) ([]map[string]interface{}, error)
	
	// GetProviderTables returns available tables for a provider
	GetProviderTables(ctx context.Context, provider CloudProvider) ([]string, error)
	
	// GetTableSchema returns the schema for a specific table
	GetTableSchema(ctx context.Context, tableName string) (map[string]string, error)
	
	// DiscoverWithSteampipe discovers resources using Steampipe queries
	DiscoverWithSteampipe(ctx context.Context, provider CloudProvider, opts ProviderDiscoveryOptions) ([]Resource, error)
}

// ResourceFilter defines criteria for filtering resources during discovery
type ResourceFilter struct {
	Type     FilterType      `json:"type"`
	Field    string          `json:"field"`
	Operator FilterOperator  `json:"operator"`
	Value    interface{}     `json:"value"`
	Values   []interface{}   `json:"values,omitempty"`
}

// FilterType represents the type of filter
type FilterType string

const (
	FilterTypeInclude FilterType = "include"
	FilterTypeExclude FilterType = "exclude"
)

// FilterOperator represents filter operators
type FilterOperator string

const (
	FilterOperatorEquals       FilterOperator = "equals"
	FilterOperatorNotEquals    FilterOperator = "not_equals"
	FilterOperatorContains     FilterOperator = "contains"
	FilterOperatorNotContains  FilterOperator = "not_contains"
	FilterOperatorStartsWith   FilterOperator = "starts_with"
	FilterOperatorEndsWith     FilterOperator = "ends_with"
	FilterOperatorRegex        FilterOperator = "regex"
	FilterOperatorIn           FilterOperator = "in"
	FilterOperatorNotIn        FilterOperator = "not_in"
	FilterOperatorGreaterThan  FilterOperator = "greater_than"
	FilterOperatorLessThan     FilterOperator = "less_than"
	FilterOperatorExists       FilterOperator = "exists"
	FilterOperatorNotExists    FilterOperator = "not_exists"
)

// CredentialProvider defines interface for providing cloud credentials
type CredentialProvider interface {
	// GetCredentials returns credentials for the specified provider
	GetCredentials(ctx context.Context, provider CloudProvider) (map[string]string, error)
	
	// ValidateCredentials validates the provided credentials
	ValidateCredentials(ctx context.Context, provider CloudProvider, credentials map[string]string) error
	
	// RefreshCredentials refreshes expired credentials if possible
	RefreshCredentials(ctx context.Context, provider CloudProvider) error
}

// Cache defines interface for caching discovery results
type Cache interface {
	// Get retrieves cached discovery results
	Get(ctx context.Context, key string) (*DiscoveryResult, error)
	
	// Set stores discovery results in cache
	Set(ctx context.Context, key string, result *DiscoveryResult, ttl time.Duration) error
	
	// Delete removes cached results
	Delete(ctx context.Context, key string) error
	
	// Clear clears all cached results
	Clear(ctx context.Context) error
	
	// Keys returns all cache keys
	Keys(ctx context.Context) ([]string, error)
}

// EventHandler defines interface for handling discovery events
type EventHandler interface {
	// OnDiscoveryStart is called when discovery starts
	OnDiscoveryStart(ctx context.Context, opts DiscoveryOptions)
	
	// OnDiscoveryComplete is called when discovery completes
	OnDiscoveryComplete(ctx context.Context, result *DiscoveryResult)
	
	// OnProviderStart is called when provider discovery starts
	OnProviderStart(ctx context.Context, provider CloudProvider)
	
	// OnProviderComplete is called when provider discovery completes
	OnProviderComplete(ctx context.Context, provider CloudProvider, resources []Resource)
	
	// OnResourceDiscovered is called when a resource is discovered
	OnResourceDiscovered(ctx context.Context, resource Resource)
	
	// OnError is called when an error occurs
	OnError(ctx context.Context, err DiscoveryError)
}
